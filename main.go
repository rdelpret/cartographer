package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
  Sources []string `yaml:"sources"`
  Destination string `yaml:"destination"`
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
func PrintApp(file string, num int) app {
	a := loadAppFile(file)
	fmt.Printf("%d.\n", num)
	fmt.Println("    sources:")

	for _, s := range a.Sources {
		fmt.Println("    - name:", s.Name)
		fmt.Println("      github:", s.Github)
		fmt.Println("      path:", s.Path)

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

func main() {
	files := getConfFiles("apps/")
	fmt.Printf("Welcome to the Cartograhper! I found %d application(s) to fetch.\n", len(files))
	fmt.Println()
	for i, f := range files {
		PrintApp(f, i+1)
	}

}
