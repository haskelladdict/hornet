// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

const version = "0.1"

// vars for command line parser
var (
	numThreads int
	hashType   string
	doStats    bool
	refFile    string
	showErrors bool
)

func init() {
	flag.IntVar(&numThreads, "n", 1, "number of analysis threads")
	flag.StringVar(&hashType, "h", "md5", "hash type: md5 (default), sha1, sha512")
	flag.BoolVar(&showErrors, "e", false, "show any errors at the end or a run")
	flag.BoolVar(&doStats, "s", false, "print final runtime statistics")
	flag.StringVar(&refFile, "c", "", "compare against <filename> containing output\n"+
		"\tfrom previous run and report changed or missing files")
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		usageAndExit()
	}
	root := flag.Arg(0)
	if hashType != "md5" && hashType != "sha1" && hashType != "sha512" {
		usageAndExit()
	}
	runtime.GOMAXPROCS(int(numThreads))

	// parse reference file if requested
	var fileMap map[string]string
	if refFile != "" {
		var err error
		fileMap, hashType, err = parseReference(refFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse reference file %s: %s\n", refFile, err)
			usageAndExit()
		}
	}

	// main processing loop
	start := time.Now()
	results := make(chan FileInfo)
	fileQueue := make(chan string)

	// done will close down all active goroutines
	done := make(chan struct{})
	defer close(done)

	go fileWalker(done, root, fileQueue, results)

	var wg sync.WaitGroup
	for i := 0; i < int(numThreads); i++ {
		wg.Add(1)
		go digester(done, fileQueue, results, hashType, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// either print hashes and errors/stats if requested or compare against
	// the reference
	var totalFileSize, numFiles int64
	var errors []error
	if refFile != "" {
		totalFileSize, numFiles, errors = compareHashes(results, fileMap)
	} else {
		totalFileSize, numFiles, errors = printHashes(results, hashType)
	}

	if doStats {
		printStats(start, numFiles, totalFileSize, len(errors))
	}

	if showErrors {
		printErrors(errors)
	}
}
