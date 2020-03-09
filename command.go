package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
)

func printContext(path string, ctx *cli.Context) {
	fmt.Printf("Running Sypht-cli with command %s, directory : %v, recursively : %v, upload rate : %v doc(s)/second", ctx.Command.Name, path, cliFlags.recursive, cliFlags.uploadRate)
	fmt.Println()
}

func watch(path string, ctx *cli.Context) error {
	if cliFlags.nThreads == 1 {
		cliFlags.nThreads = cliFlags.uploadRate
	}
	printContext(path, ctx)
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
		csvWriter.Flush()
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	paths := make(chan string)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					paths <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	var errChan <-chan error
	var dirs <-chan string
	if cliFlags.recursive {
		doneChan := make(chan struct{})
		defer close(doneChan)
		dirs, errChan = walkDirs(doneChan, path)
		for dir := range dirs {
			err = watcher.Add(dir)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Start a fixed number of goroutines to read and process files.
	processDone := make(chan struct{})
	defer close(processDone)
	c := make(chan uploadResult)
	var wg sync.WaitGroup

	wg.Add(cliFlags.nThreads)
	for i := 0; i < cliFlags.nThreads; i++ {
		go func() {
			processFile(processDone, paths, c)
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

	if err := <-errChan; err != nil {
		fmt.Printf("Walk failed %v", err)
	}

	<-done
	return nil
}

func scan(path string, ctx *cli.Context) error {
	if cliFlags.nThreads == 1 {
		cliFlags.nThreads = cliFlags.uploadRate
	}
	printContext(path, ctx)
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
		csvWriter.Flush()
	}
	done := make(chan struct{})
	defer close(done)
	paths, errChan := walkFiles(done, path)

	// Start a fixed number of goroutines to read and process files.
	c := make(chan uploadResult)
	var wg sync.WaitGroup

	wg.Add(cliFlags.nThreads)
	for i := 0; i < cliFlags.nThreads; i++ {
		go func() {
			processFile(done, paths, c)
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
	// Check whether the Walk failed.
	if err := <-errChan; err != nil {
		fmt.Printf("Walk failed %v", err)
	}
	return nil
}
