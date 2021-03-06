// +build ignore

// 本程序用来生成 app/contributors.go，无需编译
package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Contributor struct {
	Name  string
	Lines int
}

func main() {
	authorList := parseContributors(RunCommand(`git log --stat`))
	appVersion := RunCommand(`git describe --always --tags --dirty`)
	buildHost := RunCommand(`hostname`)
	goVersion := RunCommand(`go version`)

	file, err := os.Create("version.go")
	if err != nil {
		log.Fatal(`os.Create("version.go"): `, err)
		return
	}

	now := time.Now()
	fileTemplate.Execute(file, struct {
		Timestamp      time.Time
		Carls          []Contributor
		Version        string
		BuildGoVersion string
		BuildHost      string
		GoVersion      string
	}{
		Timestamp: now,
		Carls:     authorList,
		Version:   appVersion,
		BuildHost: buildHost,
		GoVersion: goVersion,
	})
}

func parseContributors(gitLog string) []Contributor {
	var author string

	authorDict := make(map[string]int)
	lines := strings.Split(gitLog, "\n")

	for _, line := range lines {
		fields := strings.SplitN(line, " ", 2)
		if fields[0] == "Author:" {
			author = fields[1]
			continue
		}

		re := regexp.MustCompile(` (\d+) insertion`)
		subs := re.FindStringSubmatch(line)
		if subs != nil {
			lines, _ := strconv.Atoi(subs[1])
			authorDict[author] += lines
		}
	}

	var authorList []Contributor

	for k, v := range authorDict {
		authorList = append(authorList, Contributor{
			Name:  k,
			Lines: v,
		})
	}

	sort.SliceStable(authorList, func(i, j int) bool {
		return authorList[i].Lines > authorList[j].Lines
	})

	return authorList
}

func RunCommand(cmdLine string) string {
	args := regexp.MustCompile(`\s+`).Split(cmdLine, -1)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()

	if err != nil {
		log.Fatal(cmdLine, ": ", err)
	}

	return strings.Trim(string(output), "\r\n\t ")
}

var fileTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at {{ .Timestamp }}
package app

var Contributors = []struct{
	Name  string
	Lines int
} {
{{- range .Carls }}
	{{ printf "{%q, %d}" .Name .Lines }},
{{- end }}
}

var (
	AppName         = "GoMud"
	Version         = {{ printf "%q" .Version }}
	BuildTime       = {{.Timestamp.Format "2006-01-02 15:04:05 MST" | printf "%q"}}
	BuildGoVersion  = {{ printf "%q" .GoVersion }}
	BuildHost       = {{ printf "%q" .BuildHost }}
)
`))
