package main

import (
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
)

type conf struct {
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


func loadConfig(file string) (c *conf) {
  source, err := ioutil.ReadFile(file)

  if err != nil {
      log.Fatal("Couldn't read yaml file:", err)
  }

  err = yaml.Unmarshal(source, &c)

  if err != nil {
      log.Fatal("Couldn't parse yaml file:", err)
  }
  return c
}


func main() {
    c := loadConfig("apps/cluster-autoscaler.yaml")
    fmt.Println(c)
}
