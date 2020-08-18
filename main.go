package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"github.com/hashicorp/go-getter"
	"time"
)

// data structure for application
type app struct {
	Destinations []destination `yaml:"destinations"`
	Sources      []source      `yaml:"sources"`
	Routes       []route       `yaml:"routes"`
}

type destination struct {
	Name   string `yaml:"name"`
	Github string `yaml:"github"`
	Path   string `yaml:"path"`
}

type source struct {
	Name   string `yaml:"name"`
	Github string `yaml:"github"`
	Path   string `yaml:"path"`
	Branch string `yaml:"branch"`
}

type route struct {
	Sources     []string `yaml:"sources"`
	Destination string   `yaml:"destination"`
	ObjectTypes []string `yaml:"objectTypes"`
}

// function that loads the config into the structs
func loadAppFile(file string) app {

	var a app

	source, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("Couldn't read yaml file:", err)
	}

	err = yaml.Unmarshal(source, &a)

	if err != nil {
		log.Fatal("Couldn't parse yaml file:", err)
	}

	return a

}

// function that pulls all the file names from
// a directory and loads them into a slice
func getConfFiles(dir string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files[1:]
}

// function to demonstrate that we can
// pull stuff in from yaml and create structs
func ProcessApp(file string, num int) app {
	a := loadAppFile(file)
	fmt.Printf("%d.\n", num)
	fmt.Println("    sources:")

	for _, s := range a.Sources {
		fmt.Println("    - name:", s.Name)
		fmt.Println("      github:", s.Github)
		fmt.Println("      path:", s.Path)
    downloadSource(s)
	}

	fmt.Println()
	fmt.Println("    destinations:")

	for _, d := range a.Destinations {
		fmt.Println("    - name:", d.Name)
		fmt.Println("      github:", d.Github)
		fmt.Println("      path:", d.Path)
	}

	fmt.Println()
	fmt.Println("    routes:")
	for _, r := range a.Routes {
		fmt.Println("    - sources:", r.Sources)
		fmt.Println("      objectTypes:", r.ObjectTypes)
		fmt.Println("      destination:", r.Destination)
	}

	fmt.Println()

	return a
}

func githubStuff(token string) string {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	repos, _, err := client.Repositories.List(ctx, "", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(repos)
	return token
}

func loadGithubToken() string {

	token, err := ioutil.ReadFile("secret/token")

	if err != nil {
		log.Fatal("Couldn't read token file:", err)
	}

	return strings.TrimSuffix(string(token), "\n")
}

func downloadSource(s source) string {
	url := "https://github.com/" + s.Github + "//" + s.Path
	loc := "repos/" + s.Name
	err := getter.Get(loc, "git::" + url + "?ref=" + s.Branch)
	if err != nil {
	   log.Fatal(err)
	}
	return loc
}

func main() {
	//token := loadGithubToken()

	files := getConfFiles("apps/")
	fmt.Printf("Welcome to the Cartograhper! I found %d application(s) to fetch.\n", len(files))
	fmt.Println()
	for i, f := range files {
		ProcessApp(f, i+1)
  }

	time.Sleep(5 * time.Second)

	err := os.RemoveAll("repos/")
    if err != nil {
        log.Fatal(err)
    }
		
  //githubStuff(token)

}
