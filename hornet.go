// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
)

func init() {
	flag.IntVar(&numThreads, "n", 1, "number of analysis threads")
	flag.StringVar(&hashType, "h", "md5", "hash type: md5 (default), sha1, sha512")
	flag.BoolVar(&doStats, "s", false, "print final runtime statistics")
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		usage()
		os.Exit(1)
	}
	root := flag.Arg(0)
	if hashType != "md5" && hashType != "sha1" && hashType != "sha256" {
		usage()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(int(numThreads))

	// main processing loop
	start := time.Now()
	fileQueue := make(chan string)
	go fileWalker(root, fileQueue)

	var totalFileSize int64
	var numFiles int64
	var wg sync.WaitGroup
	for i := 0; i < int(numThreads); i++ {
		wg.Add(1)
		go func() {
			for f := range fileQueue {
				hashString, size, err := hasher(f)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}
				numFiles += 1
				totalFileSize += int64(size)
				fmt.Printf("MD5, %s, %s\n", f, hashString)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if doStats {
		printStats(start, numFiles, totalFileSize)
	}
}

// fileWalker walks the directory tree starting at root and adds all
// encountered file paths to the file queue
// NOTE: Currently we only track regular files and symbolic links
func fileWalker(root string, fileQueue chan<- string) error {

	defer close(fileQueue)

	walker := func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			// log errors and skip over them
			fmt.Fprintln(os.Stderr, err)
			return nil
		}
		if m := info.Mode(); (m&os.ModeType == 0) || (m&os.ModeSymlink != 0) {
			fileQueue <- pth
		}
		return nil
	}

	if err := filepath.Walk(root, walker); err != nil {
		return err
	}
	return nil
}

// hasher computes the requested hash for the file at filepath. It also returns
// the total file size for runtime statistics
// NOTE: It appears that the buffer size has to be sizable in order to get
// file parsing and hashing. Based on benchmarking 1 MB buffer sizes provide
// good throughput.
func hasher(filePath string) (string, int, error) {
	fp, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	buffer := make([]byte, 1000000)
	var fileSize int
	h := md5.New()
	for {
		c, _ := reader.Read(buffer)
		fileSize += c
		if c == 0 {
			break // done
		}
		_, err := h.Write(buffer[:c])
		if err != nil {
			return "", 0, err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), fileSize, nil
}

//
