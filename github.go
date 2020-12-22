package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type pullRequest struct {
	sourceOwner   string
	sourceRepo    string
	commitMessage string
	commitBranch  string
	baseBranch    string
	prRepoOwner   string
	prRepo        string
	prBranch      string
	prSubject     string
	prDescription string
	sourceFiles   string
	authorName    string
	authorEmail   string
}

var client *github.Client
var ctx = context.Background()

func getRef(pr pullRequest, client *github.Client) (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(ctx, pr.sourceOwner, pr.sourceRepo, "refs/heads/"+pr.commitBranch); err == nil {
		return ref, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(ctx, pr.sourceOwner, pr.sourceRepo, "refs/heads/"+pr.baseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + pr.commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(ctx, pr.sourceOwner, pr.sourceRepo, newRef)
	return ref, err
}

func getTree(pr pullRequest, client *github.Client, ref *github.Reference) (tree *github.Tree, err error) {
	// Create a tree with what to commit.
	entries := []*github.TreeEntry{}

	// Load each file into the tree.
	for _, fileArg := range strings.Split(pr.sourceFiles, ",") {
		file, content, err := getFileContent(fileArg)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &github.TreeEntry{Path: github.String(file), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}

	tree, _, err = client.Git.CreateTree(ctx, pr.sourceOwner, pr.sourceRepo, *ref.Object.SHA, entries)
	return tree, err
}

// getFileContent loads the local content of a file and return the target name
// of the file in the target repository and its contents.
func getFileContent(fileArg string) (targetName string, b []byte, err error) {
	var localFile string
	files := strings.Split(fileArg, ":")
	switch {
	case len(files) < 1:
		return "", nil, errors.New("empty `-files` parameter")
	case len(files) == 1:
		localFile = files[0]
		targetName = files[0]
	default:
		localFile = files[0]
		targetName = files[1]
	}

	b, err = ioutil.ReadFile(localFile)
	return targetName, b, err
}

func pushCommit(pr pullRequest, client *github.Client, ref *github.Reference, tree *github.Tree) (err error) {
	// Get the parent commit to attach the commit to.
	parent, _, err := client.Repositories.GetCommit(ctx, pr.sourceOwner, pr.sourceRepo, *ref.Object.SHA)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: github.String(pr.authorName), Email: github.String(pr.authorEmail)}
	commit := &github.Commit{Author: author, Message: github.String(pr.commitMessage), Tree: tree, Parents: []*github.Commit{parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(ctx, pr.sourceOwner, pr.sourceRepo, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(ctx, pr.sourceOwner, pr.sourceRepo, ref, false)
	return err
}

func createPR(pr pullRequest, client *github.Client) (err error) {
	if pr.prSubject == "" {
		return errors.New("missing `-pr-title` flag; skipping PR creation")
	}

	if pr.prRepoOwner != "" && pr.prRepoOwner != pr.sourceOwner {
		pr.commitBranch = fmt.Sprintf("%s:%s", pr.sourceOwner, pr.commitBranch)
	} else {
		pr.prRepoOwner = pr.sourceOwner
	}

	if pr.prRepo == "" {
		pr.prRepo = pr.sourceRepo
	}

	newPR := &github.NewPullRequest{
		Title:               github.String(pr.prSubject),
		Head:                github.String(pr.commitBranch),
		Base:                github.String(pr.prBranch),
		Body:                github.String(pr.prDescription),
		MaintainerCanModify: github.Bool(true),
	}

	preq, _, err := client.PullRequests.Create(ctx, pr.prRepoOwner, pr.prRepo, newPR)
	if err != nil {
		return err
	}

	log.Printf("PR created: %s\n", preq.GetHTMLURL())
	return nil
}

func checkPR(pr pullRequest, client *github.Client) (bool, error) {

	if pr.prRepoOwner != "" && pr.prRepoOwner != pr.sourceOwner {
		pr.commitBranch = fmt.Sprintf("%s:%s", pr.sourceOwner, pr.commitBranch)
	} else {
		pr.prRepoOwner = pr.sourceOwner
	}

	if pr.prRepo == "" {
		pr.prRepo = pr.sourceRepo
	}

	prs, _, err := client.PullRequests.List(ctx, pr.prRepoOwner, pr.prRepo, nil)
	if err != nil {
		return false, err
	}
	for _, p := range prs {
		if *p.Title == pr.prSubject {
			return true, err
		}
	}
	return false, err
}

func loadGithubToken() string {

	token, err := ioutil.ReadFile("secret/token")

	if err != nil {
		log.Fatal("Couldn't read token file:", err)
	}

	return strings.TrimSuffix(string(token), "\n")
}

func makePR(pr pullRequest) {


	token := loadGithubToken()

	if token == "" {
		log.Fatal("No github token found")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	exists, err := checkPR(pr, client)
	if err != nil {
		log.Printf("Unable to list open PRs: %s\n", err)
		return
	}
	if exists == true {
		log.Printf("PR already exists, skipping.")
		return
	}
	ref, err := getRef(pr, client)
	if err != nil {
		log.Printf("Unable to get/create the commit reference: %s\n", err)
		return
	}
	if ref == nil {
		log.Printf("No error where returned but the reference is nil")
		return
	}

	tree, err := getTree(pr, client, ref)
	if err != nil {
		log.Printf("Unable to create the tree based on the provided files: %s\n", err)
		return
	}

	if err := pushCommit(pr, client, ref, tree); err != nil {
		log.Printf("Unable to create the commit: %s\n", err)
		return
	}

	if err := createPR(pr, client); err != nil {
		log.Printf("Error while creating the pull request: %s", err)
		return
	}

}
