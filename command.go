package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/urfave/cli/v2"
)

func watch(path string, ctx *cli.Context) error {
	csvPath := filepath.Join(path, "sypht.csv")
	exist := true
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		exist = false
	}
	metaFile, err := os.OpenFile(csvPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error creating meta file , %v", err)
	}
	defer metaFile.Close()

	csvWriter = csv.NewWriter(metaFile)
	if !exist {
		csvWriter.Write([]string{"FileId", "Path", "Status", "UploadedAt", "Error", "Checksum"})
	}

	done := make(chan struct{})
	defer close(done)
	paths, errChan := walkFiles(done, currentDir)

	// Start a fixed number of goroutines to read and process files.
	c := make(chan uploadResult)
	var wg sync.WaitGroup

	wg.Add(cliFlags.nRoutines)
	for i := 0; i < cliFlags.nRoutines; i++ {
		go func() {
			processFile(done, paths, c, 1)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	for r := range c {
		log.Printf("%v", r)
	}
	csvWriter.Flush()
	// Check whether the Walk failed.
	if err := <-errChan; err != nil {
		fmt.Printf("Walk failed %v", err)
	}
	return nil
}
