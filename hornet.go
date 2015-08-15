// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"flag"
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
	showErrors bool
)

func init() {
	flag.IntVar(&numThreads, "n", 1, "number of analysis threads")
	flag.StringVar(&hashType, "h", "md5", "hash type: md5 (default), sha1, sha512")
	flag.BoolVar(&showErrors, "e", false, "show any errors at the end or a run")
	flag.BoolVar(&doStats, "s", false, "print final runtime statistics")
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		usage()
		os.Exit(1)
	}
	root := flag.Arg(0)
	if hashType != "md5" && hashType != "sha1" && hashType != "sha512" {
		usage()
		os.Exit(1)
	}
	runtime.GOMAXPROCS(int(numThreads))

	// main processing loop
	start := time.Now()
	results := make(chan FileInfo)
	fileQueue := make(chan string)
	go fileWalker(root, fileQueue, results)

	var wg sync.WaitGroup
	for i := 0; i < int(numThreads); i++ {
		wg.Add(1)
		go digester(fileQueue, results, hashType, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// print hashes and errors/stats if requested
	totalFileSize, numFiles, errors := printHashes(results, hashType)

	if doStats {
		printStats(start, numFiles, totalFileSize, len(errors))
	}

	if showErrors {
		printErrors(errors)
	}
}

//
