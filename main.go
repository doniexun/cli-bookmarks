// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package main provides a utility to quickly navigate to bookmarked directories on the command line.
*/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/cli-bookmarks/completion"
	"github.com/DavidGamba/cli-bookmarks/config"
	"github.com/DavidGamba/cli-bookmarks/gui"
	"github.com/DavidGamba/cli-bookmarks/semver"
	"github.com/DavidGamba/go-getoptions"
)

func synopsis() {
	synopsis := `# Show the GUI
cb

# Use a bookmark
cb bookmark_alias

# Use a bookmark and navigate its subdirectories
cb bookmark_alias <tab>

Bookmarks are stored in the ~/.cli-bookmarks.toml file.

# Show help
cli-bookmarks --help

# Show version
cli-bookmarks --version
`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {
	log.SetOutput(ioutil.Discard)
	var completionCur, completionPrev string
	opt := getoptions.New()
	opt.Bool("help", false)
	opt.Bool("debug", false)
	opt.Bool("version", false)
	opt.StringVarOptional(&completionCur, "completion-current", "")
	opt.StringVarOptional(&completionPrev, "completion-previous", "")
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("help") {
		synopsis()
		os.Exit(1)
	}
	if opt.Called("version") {
		fmt.Println(semver.Version{Version: "0.1.0"})
		os.Exit(1)
	}
	if opt.Called("debug") {
		log.SetOutput(os.Stderr)
	}
	log.Printf("Called with: %v", os.Args)
	log.Printf("remaining: %v", remaining)

	cfg, err := readConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR Failed to read config file: %s\n", err)
		os.Exit(1)
	}
	log.Printf("Config: %v\n", cfg)
	// Command line completion in use
	if opt.Called("completion-current") {
		dirList, err := completion.CompletionResults(cfg.Bookmarks, completionCur)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		log.Printf("Cur: %s, Prev: %s, remaining: %v, dirList: %#v\n", completionCur, completionPrev, remaining, dirList)
		log.Printf("Output: ' %s '\n", strings.Join(dirList, " "))
		fmt.Printf(" %s ", strings.Join(dirList, " "))
		os.Exit(0)
	}
	if len(remaining) != 0 {
		alias := remaining[0]
		parts := []string{}
		if strings.Contains(alias, "/") {
			parts = strings.SplitN(alias, "/", 2)
			alias = parts[0]
		}
		if v, ok := cfg.Bookmarks[alias]; ok {
			if len(parts) >= 2 {
				log.Println(v + "/" + parts[1])
				fmt.Println(v + "/" + parts[1])
			} else {
				log.Println(v)
				fmt.Println(v)
			}
			os.Exit(0)
		}
		// TODO: Show this in the interactive window?
		fmt.Fprintf(os.Stderr, "Wrong alias: %s\n", remaining[0])
		os.Exit(1)
	}
	bgui := gui.New(cfg.Bookmarks)
	if opt.Called("debug") {
		bgui.Debug = true
	}
	err = bgui.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}

func readConfigFile() (config.Config, error) {
	configFilePath := os.Getenv("HOME") + string(os.PathSeparator) + ".cli-bookmarks.toml"
	return config.ParseFile(configFilePath)
}
