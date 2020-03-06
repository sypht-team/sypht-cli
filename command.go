package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
)

func watch(path string, ctx *cli.Context) error {
	initCSV(path)
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

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
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
	csvWriter.Flush()

	<-done
	return nil
}

func scan(path string, ctx *cli.Context) error {
	initCSV(path)
	done := make(chan struct{})
	defer close(done)
	paths, errChan := walkFiles(done, currentDir)

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
	csvWriter.Flush()
	// Check whether the Walk failed.
	if err := <-errChan; err != nil {
		fmt.Printf("Walk failed %v", err)
	}
	return nil
}
