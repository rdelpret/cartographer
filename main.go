package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/go-getter"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// --- DATA DEFINITIONS ---

type app struct {
	Destinations []destination `yaml:"destinations"`
	Sources      []source      `yaml:"sources"`
	Routes       []route       `yaml:"routes"`
}

type source struct {
	Name   string   `yaml:"name"`
	Github string   `yaml:"github"`
	Path   string   `yaml:"path"`
	Files  []string `yaml:files`
	Branch string   `yaml:"branch"`
}

type destination struct {
	Name   string `yaml:"name"`
	Github string `yaml:"github"`
	Path   string `yaml:"path"`
}

type route struct {
	Sources     []string `yaml:"sources"`
	Destination string   `yaml:"destination"`
	ObjectTypes []string `yaml:"objectTypes"`
}

// --- UTILITY FUNCTIONS ---

// simple function to test if string exists in []string
func contains(l []string, s string) bool {
	for _, a := range l {
		if a == s {
			return true
		}
	}
	return false
}

func hash(file string) string {

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))[0:7]
}

//deletes a given directory
func cleanup(dir string) {
	log.Println("cleaning up", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Println(err)
	}
	return
}

// function that pulls all the file names from
// a directory and loads them into a slice
func ls(dir string) []string {
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

// --- FILE SYSTEM / YAML DATA FUNCTIONS ---

// loads the config file into structs
// this will later become a k8s CRD
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

// Downloads the source repo and places it in a temp directory
func downloadSource(s source) {
	url := "https://github.com/" + s.Github + "//" + s.Path
	loc := "temp/sources/" + s.Name
	err := getter.Get(loc, "git::"+url+"?ref="+s.Branch)
	if err != nil {
		fmt.Println(err)
	}
	return
}

// Downloads a destination repo and puts in a temp directory
func downloadDestination(d destination) []string {
	url := "https://github.com/" + d.Github + "//" + d.Path
	loc := "temp/destinations/" + d.Name
	err := getter.Get(loc, "git::"+url+"?ref=master")
	if err != nil {
		fmt.Println(err)
	}
	return ls(loc)
}

// --- CARTOPGRAPHER APP LOGIC ---

// Main function that process an "app"
func processApp(file string) {

	log.Print("Loading ", file)
	a := loadAppFile(file)
	var sourceFileListFromYaml []string
	var destFileListFromDir []string
	var destGitHub []string
	var sourceGitHub []string

	// all processing of sources that do not need awareness of destinations
	for _, s := range a.Sources {
		log.Printf("Processing Source {name: %s, github: %s, path: %s}\n", s.Name, s.Github, s.Path)
		log.Println("Downloading source from", "https://github.com/"+s.Github+"//"+s.Path)

		downloadSource(s)

		for _, f := range s.Files {
			sourceFileListFromYaml = append(sourceFileListFromYaml, "temp/sources/"+s.Name+"/"+f)
		}
		sourceGitHub = strings.Split(s.Github, "/")
	}
	// all processing of destinations that do not need awareness of sources
	for _, d := range a.Destinations {
		log.Printf("Processing destination {name: %s, github: %s, path: %s}\n", d.Name, d.Github, d.Path)
		log.Println("Downloading destination from", "https://github.com/"+d.Github+"//"+d.Path)

		downloadedFileList := downloadDestination(d)

		for _, f := range downloadedFileList {
			destFileListFromDir = append(destFileListFromDir, "temp/destinations/"+d.Name+"/"+f)
		}
		destGitHub = strings.Split(d.Github, "/")
	}

	// For now assume the file list for each source must appear in each destination
	// we will need to do some work here to impliment routes but that can be plugged in later
	sourceFiles := ""

	for _, f := range sourceFileListFromYaml {
		if !(contains(destFileListFromDir, f)) {
			if sourceFiles == "" {
				sourceFiles = sourceFiles + f
			} else {
				sourceFiles = sourceFiles + "," + f
			}

		} else {
			log.Println("I must diff file", f)
		}

	}
    fileHash := hash(sourceFiles)
	var pr pullRequest
	pr.sourceOwner = destGitHub[0]
	pr.sourceRepo = destGitHub[1]
	pr.commitMessage = "Cartographer: Update " + sourceGitHub[1]
	pr.commitBranch = "cartographer/" + sourceGitHub[1] + "/" + fileHash
	pr.baseBranch = "master"
	pr.prRepoOwner = destGitHub[0]
	pr.prRepo = destGitHub[1]
	pr.prBranch = "master"
	pr.prSubject = "Cartographer: Update " + sourceGitHub[1] + " [" + fileHash + "]"
	pr.prDescription = "Cartographer: Update " + sourceGitHub[1]
	pr.sourceFiles = sourceFiles
	pr.authorName = "rdelpret"
	pr.authorEmail = "robbie@lola.com"

	makePR(pr)

	return
}

// --- MAIN LOOP ---

func main() {
	defer cleanup("temp/")
	for {
		files := ls("apps/")
		log.Printf("Found %d application(s) in /app to process\n", len(files))
		for _, f := range files {
			processApp(f)
		}
		time.Sleep(60 * time.Second)

		cleanup("temp/")
	}

}
