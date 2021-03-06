// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"path/filepath"
)

var cmdClean = &command{
	run:   runClean,
	Name:  "clean",
	Usage: "",
	Short: fmt.Sprintf("remove object files and cached %s files", gomobileName),
	Long: fmt.Sprintf(`
Clean removes object files and cached NDK files downloaded by %s init
`, gomobileName),
}

func runClean(cmd *command) (err error) {
	gopaths := filepath.SplitList(goEnv("GOPATH"))
	if len(gopaths) == 0 {
		return fmt.Errorf("GOPATH is not set")
	}
	gomobilepath = filepath.Join(gopaths[0], "pkg/"+gomobileName)
	if buildX {
		fmt.Fprintln(xout, gomobileEnvName+"="+gomobilepath)
	}
	return removeAll(gomobilepath)
}
