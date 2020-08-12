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
  Destination destination `yaml:"destination"`
  Source  source `yaml:"source"`
}

type destination struct {
  Name string `yaml:"name"`
  Github string `yaml:"github"`
  Path string `yaml:"path"`
}

type source struct {
  Name string `yaml:"name"`
  Github string `yaml:"github"`
  Path string `yaml:"path"`
  Branch string `yaml:"branch"`
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
func PrintApp (file string, num int) app {
  a := loadAppFile(file)
  fmt.Printf("%d.\n", num)
  fmt.Println("    ---- SOURCE ----")
  fmt.Println("    source.name:", a.Source.Name)
  fmt.Println("    source.github:", a.Source.Github)
  fmt.Println("    source.path:", a.Source.Path)
  fmt.Println()
  fmt.Println("    -- DESTINATION --")
  fmt.Println("    destination.name:", a.Destination.Name)
  fmt.Println("    destination.name:", a.Destination.Github)
  fmt.Println("    destination.path:", a.Destination.Path)
  fmt.Println()
  return a
}

func main() {
    files:= getConfFiles("apps/")
    fmt.Printf("Welcome to the Cartograhper! I found %d application(s) to fetch.\n", len(files))
    fmt.Println()
    for i, f := range files {
      PrintApp(f, i + 1)
    }

}
