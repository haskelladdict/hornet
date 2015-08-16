// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// printHashes prints the file paths and corresponding hashes to stdout. It
// also returns the total number of files and total file size processed
func printHashes(results <-chan FileInfo, hashType string) (int64, int64, []error) {
	var totalFileSize int64
	var numFiles int64
	var errors []error
	for info := range results {
		if info.Error != nil {
			errors = append(errors, info.Error)
			continue
		}
		numFiles++
		totalFileSize += int64(info.Size)
		fmt.Printf("%s,\"%s\",%s\n", hashType, info.Path, info.Hash)
	}
	return totalFileSize, numFiles, errors
}

// printStats prints runtime statistics
func printStats(start time.Time, numFiles int64, totalFileSize int64, numErrors int) {
	fileSizeMB := totalFileSize / 1024 / 1024
	execTime := time.Since(start).Seconds()
	fmt.Println()
	fmt.Println("--------------  RUNTIME STATS -----------------")
	fmt.Printf("hornet version   : %s\n", version)
	fmt.Printf("date             : %s\n", time.Now().Format(time.RFC1123))
	fmt.Printf("elapsed time     : %f s\n", execTime)
	fmt.Printf("# file errors    : %d\n", numErrors)
	fmt.Printf("files processed  : %d\n", numFiles)
	fmt.Printf("data processed   : %d MB\n", fileSizeMB)
	fmt.Printf("throughput       : %f MB/s\n", float64(fileSizeMB)/execTime)
}

// printErrors prints any encountered runtime errors
func printErrors(errors []error) {
	fmt.Println()
	fmt.Println("--------------  RUNTIME ERRORS -----------------")
	for i, e := range errors {
		fmt.Println(i, e)
	}
}

// usage prints a brief usage info and then exits with status 1
func usageAndExit() {
	fmt.Printf("hornet v%s  (C) 2015 Markus Dittrich", version)
	fmt.Println()
	fmt.Println("usage: hornet <options> <directory root or filename>")
	fmt.Println()
	fmt.Println("options:")
	flag.PrintDefaults()
	os.Exit(1)
}
