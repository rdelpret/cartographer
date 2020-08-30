package main

import (
	//"context"
	"fmt"
	//"github.com/google/go-github/github"
	//"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	//"strings"
	"github.com/hashicorp/go-getter"
	"time"
)

// data structure for application
type app struct {
	Destinations []destination `yaml:"destinations"`
	Sources      []source      `yaml:"sources"`
	Routes       []route       `yaml:"routes"`
}

type source struct {
	Name   string `yaml:"name"`
	Github string `yaml:"github"`
	Path   string `yaml:"path"`
	Files  []string `yaml:files`
	Branch string `yaml:"branch"`
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
func downloadDestination(d destination) {
	url := "https://github.com/" + d.Github + "//" + d.Path
	loc := "temp/destinations/" + d.Name
	err := getter.Get(loc, "git::"+url+"?ref=master")
	if err != nil {
		fmt.Println(err)
	}
	return
}

// Main function that process an "app"
func ProcessApp(file string, num int) {

	log.Print("Loading ", file)
	a := loadAppFile(file)

	for _, s := range a.Sources {
		log.Printf("Processing Source {name: %s, github: %s, path: %s}\n", s.Name, s.Github, s.Path)
		log.Println("Downloading source from", "https://github.com/"+s.Github+"//"+s.Path)
		downloadSource(s)
	}

	for _, d := range a.Destinations {
		log.Printf("Processing destination {name: %s, github: %s, path: %s}\n", d.Name, d.Github, d.Path)
		log.Println("Downloading destination from", "https://github.com/"+d.Github+"//"+d.Path)
		downloadDestination(d)
	}
	return
}

//cleans up a directory
func cleanup(dir string) {
	log.Println("cleaning up", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func main() {
	defer cleanup("temp/")
	for {
		files := getConfFiles("apps/")
		log.Printf("Found %d application(s) in /app to process\n", len(files))
		for i, f := range files {
			ProcessApp(f, i+1)
		}
		cleanup("temp/")
		time.Sleep(5 * time.Second)
	}

}
