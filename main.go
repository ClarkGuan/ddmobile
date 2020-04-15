// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate gomobile help documentation doc.go

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	gomobileName    = "gomobile"
	gomobileEnvName = "GOMOBILE"
	goVersionOut    = []byte(nil)
)

func printUsage(w io.Writer) {
	bufw := bufio.NewWriter(w)
	if err := usageTmpl.Execute(bufw, commands); err != nil {
		panic(err)
	}
	bufw.Flush()
}

func main() {
	gomobileName = os.Args[0]
	flag.Usage = func() {
		printUsage(os.Stderr)
		os.Exit(2)
	}
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
	}

	if args[0] == "help" {
		if len(args) == 3 && args[1] == "documentation" {
			helpDocumentation(args[2])
			return
		}
		help(args[1:])
		return
	}

	if err := determineGoVersion(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", gomobileName, err)
		os.Exit(1)
	}

	for _, cmd := range commands {
		if cmd.Name == args[0] {
			cmd.flag.Usage = func() {
				cmd.usage()
				os.Exit(1)
			}
			cmd.flag.Parse(args[1:])
			if err := cmd.run(cmd); err != nil {
				msg := err.Error()
				if msg != "" {
					fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
				}
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "%s: unknown subcommand %q\nRun '%s help' for usage.\n", os.Args[0], args[0], gomobileName)
	os.Exit(2)
}

func determineGoVersion() error {
	goVersionOut, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("'go version' failed: %v, %s", err, goVersionOut)
	}
	var minor int
	if _, err := fmt.Sscanf(string(goVersionOut), "go version go1.%d", &minor); err != nil {
		// Ignore unknown versions; it's probably a devel version.
		return nil
	}
	if minor < 10 {
		return errors.New("Go 1.10 or newer is required")
	}
	return nil
}

func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return // succeeded at helping
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s help command\n\nToo many arguments given.\n", gomobileName)
		os.Exit(2) // failed to help
	}

	arg := args[0]
	for _, cmd := range commands {
		if cmd.Name == arg {
			cmd.usage()
			return // succeeded at helping
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run '%s help'.\n", arg, gomobileName)
	os.Exit(2)
}

const documentationHeader = `// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code generated by '%s help documentation doc.go'. DO NOT EDIT.
`

func helpDocumentation(path string) {
	w := new(bytes.Buffer)
	//w.WriteString(documentationHeader)
	w.WriteString(fmt.Sprintf(documentationHeader, gomobileName))
	w.WriteString("\n/*\n")
	if err := usageTmpl.Execute(w, commands); err != nil {
		log.Fatal(err)
	}

	for _, cmd := range commands {
		r, rlen := utf8.DecodeRuneInString(cmd.Short)
		w.WriteString("\n\n")
		w.WriteRune(unicode.ToUpper(r))
		w.WriteString(cmd.Short[rlen:])
		w.WriteString("\n\nUsage:\n\n\t" + gomobileName + " " + cmd.Name)
		if cmd.Usage != "" {
			w.WriteRune(' ')
			w.WriteString(cmd.Usage)
		}
		w.WriteRune('\n')
		w.WriteString(cmd.Long)
	}

	w.WriteString("*/\npackage main // import \"golang.org/x/mobile/cmd/" + gomobileName + "\"\n")

	if err := ioutil.WriteFile(path, w.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}
}

var commands = []*command{
	// TODO(crawshaw): cmdRun
	cmdBind,
	cmdBuild,
	cmdClean,
	cmdInit,
	cmdInstall,
	cmdVersion,
}

type command struct {
	run   func(*command) error
	flag  flag.FlagSet
	Name  string
	Usage string
	Short string
	Long  string
}

func (cmd *command) usage() {
	fmt.Fprintf(os.Stdout, "usage: %s %s %s\n%s", gomobileName, cmd.Name, cmd.Usage, cmd.Long)
}

func titleCase(s string) string {
	return fmt.Sprintf(s, strings.Title(gomobileName), gomobileName, gomobileName, gomobileName, gomobileName)
}

var usageTmpl = template.Must(template.New("usage").Parse(titleCase(
	`%s is a tool for building and running mobile apps written in Go.

To install:

	$ go get golang.org/x/mobile/cmd/%s
	$ %s init

At least Go 1.10 is required.
For detailed instructions, see https://golang.org/wiki/Mobile.

Usage:

	%s command [arguments]

Commands:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use '%s help [command]' for more information about that command.
`)))
