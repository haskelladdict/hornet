// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// parseReference parses the content of a hornet generate reference file
// containing previously computed file hashes. It returns a map from file
// names to checksums as well as the hash type used.
func parseReference(refFileName string) (map[string]string, string, error) {
	fileMap := make(map[string]string)
	file, err := os.Open(refFileName)
	if err != nil {
		return fileMap, "", err
	}

	reader := csv.NewReader(file)
	out, err := reader.ReadAll()
	if err != nil {
		return fileMap, "", err
	}

	hashType = ""
	for _, elem := range out {
		hashType = strings.TrimSpace(elem[0])
		fileMap[strings.TrimSpace(elem[1])] = strings.TrimSpace(elem[2])
	}

	return fileMap, hashType, nil
}

// compareHashes compares a list of hashes against a list of reference hashes
// parsed from the commandline (presumably from a previous invocation of hornet)
// We print out all files with mismatched hashes as well as files that are
// missing or new
func compareHashes(results <-chan FileInfo, refMap map[string]string) (int64,
	int64, []error) {

	fmt.Println("--------------  HASH MISMATCHES -----------------")

	var numFiles, totalFileSize int64
	var missing []string
	var errors []error
	for info := range results {
		if info.Error != nil {
			errors = append(errors, info.Error)
			continue
		}
		numFiles++
		totalFileSize += int64(info.Size)
		if refHash, ok := refMap[info.Path]; !ok {
			missing = append(missing, info.Path)
		} else {
			if refHash != info.Hash {
				fmt.Printf("%s: expected (%s)  actual (%s)\n",
					info.Path, refHash, info.Hash)
			}
			refMap[info.Path] = ""
		}
	}

	fmt.Println("\n--------------  EXTRA FILES -----------------")
	for _, f := range missing {
		fmt.Println(f)
	}

	fmt.Println("\n--------------  MISSING FILES -----------------")
	for f, v := range refMap {
		if v != "" {
			fmt.Println(f)
		}
	}

	return totalFileSize, numFiles, errors
}

//
