// Copyright 2015 Markus Dittrich
// Licensed under BSD license, see LICENSE file for details

// hornet is a tool for checksumming files in a directory tree
// and comparing against a known list of checksums.

package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"os"
	"sync"
)

// digester takes files from the file queues, calls the hash function on them
// and adds the resulting hash to the results channel
func digester(fileQueue <-chan string, results chan<- FileInfo, hashType string,
	wg *sync.WaitGroup) {
	// select hash
	var hashFn func() hash.Hash
	switch hashType {
	case "md5":
		hashFn = md5.New
	case "sha1":
		hashFn = sha1.New
	case "sha512":
		hashFn = sha512.New
	default:
		panic("unknown hash function " + hashType)
	}

	for file := range fileQueue {
		hashString, size, err := hasher(file, hashFn)
		if err != nil {
			results <- FileInfo{file, err, "", 0}
			continue
		}
		results <- FileInfo{file, nil, hashString, size}
	}
	wg.Done()
}

// hasher computes the requested hash for the file at filepath. It also returns
// the total file size for runtime statistics
// NOTE: It appears that the buffer size has to be sizable in order to get
// file parsing and hashing. Based on benchmarking 1 MB buffer sizes provide
// good throughput.
func hasher(filePath string, hashFn func() hash.Hash) (string, int, error) {
	fp, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	buffer := make([]byte, 10000)
	var fileSize int
	h := hashFn()
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
