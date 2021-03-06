// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileInfo contains the size and checksum for each examined file (or
// the error encountered during reading or computing the hash)
type FileInfo struct {
	Path  string
	Error error // processing error or nil otherwise
	Hash  string
	Size  int
}

// fileWalker walks the directory tree starting at root and adds all
// encountered file paths to the file queue
// NOTE: Currently we only track regular files and symbolic links
func fileWalker(done <-chan struct{}, root string, fileQueue chan<- string,
	results chan<- FileInfo) error {

	defer close(fileQueue)

	walker := func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			results <- FileInfo{pth, err, "", 0}
			return nil
		}
		if m := info.Mode(); m&os.ModeType == 0 || (m&os.ModeSymlink != 0) {
			select {
			case fileQueue <- pth:
			case <-done:
				return fmt.Errorf("walk canceled")
			}
		}
		return nil
	}

	if err := filepath.Walk(root, walker); err != nil {
		return err
	}
	return nil
}
