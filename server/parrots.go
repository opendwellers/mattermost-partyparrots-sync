package main

import (
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	baseURL = "https://raw.githubusercontent.com/jmhobbs/cultofthepartyparrot.com/main/"
)

var parrotTypes = [...]string{
	"parrots",
	"flags",
	"guests",
}

type list map[string]string

type Parrot struct {
	name string
	file string
	gif  []byte
}

func fetchParrotList(parrotType string) (parrots []Parrot, err error) {
	resp, err := http.Get(baseURL + parrotType + ".yaml")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var parrotList []list
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(body, &parrotList)
	if err != nil {
		return
	}

	for i := 0; i < len(parrotList); i++ {
		var parrot Parrot
		if parrotList[i]["gif"] != "" {
			parrot.file = parrotList[i]["gif"]
		} else if parrotList[i]["hd"] != "" {
			parrot.file = parrotList[i]["hd"]
		}
		split := strings.SplitAfter(parrot.file, "/")
		parrot.name, _ = strings.CutSuffix(split[len(split)-1], ".gif")
		parrots = append(parrots, parrot)
	}
	return
}

func fetchParrotGif(parrot *Parrot, parrotType string) (err error) {
	// Fetch the gif from GitHub
	resp, err := http.Get(baseURL + parrotType + "/" + parrot.file)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	parrot.gif, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}
