// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"flag"
	"fmt"
	"time"
)

// usage prints a brief usage info
func usage() {
	fmt.Printf("hornet v%s  (C) 2015 Markus Dittrich", version)
	fmt.Println()
	fmt.Println("usage: hornet <options> <directory root or filename>")
	fmt.Println()
	fmt.Println("options:")
	flag.PrintDefaults()
}

// printStats prints runtime statistics
func printStats(start time.Time, numFiles int64, totalFileSize int64) {
	fileSizeMB := totalFileSize / 1024 / 1024
	execTime := time.Since(start).Seconds()
	fmt.Println()
	fmt.Println("--------------  RUNTIME STATS -----------------")
	fmt.Printf("hornet version   : %s\n", version)
	fmt.Printf("date             : %s\n", time.Now().Format(time.RFC1123))
	fmt.Printf("elapsed time     : %f s\n", execTime)
	fmt.Printf("files processed  : %d\n", numFiles)
	fmt.Printf("data processed   : %d MB\n", fileSizeMB)
	fmt.Printf("throughput       : %f MB/s\n", float64(fileSizeMB)/execTime)
}
