package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sypht-team/sypht-golang-client"
)

var client *sypht.Client

func main() {

	currentDirectory, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	config := getConfig(filepath.Join(currentDirectory, "config.json"))
	client, err = sypht.NewSyphtClient(fmt.Sprintf("%s:%s", config.ClientID, config.ClientSecret), nil)
	if err != nil {
		log.Fatalf("Unable to start Sypht client , %v", err)
	}

	csvPath := filepath.Join(currentDirectory, "sypht.csv")
	exist := true
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		exist = false
	}

	metaFile, err := os.OpenFile(csvPath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("Error creating meta file , %v", err)
	}
	defer metaFile.Close()

	csvWriter = csv.NewWriter(metaFile)
	if !exist {
		csvWriter.Write([]string{"FileId", "Path", "Status", "UploadedAt", "", "Checksum"})
		csvWriter.Flush()
	}
	done := make(chan struct{})
	defer close(done)

	paths, errChan := walkFiles(done, currentDirectory)

	// Start a fixed number of goroutines to read and process files.
	c := make(chan uploadResult)
	var wg sync.WaitGroup
	const nRoutines = 2
	wg.Add(nRoutines)
	for i := 0; i < nRoutines; i++ {
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

	// Check whether the Walk failed.
	if err := <-errChan; err != nil {
		fmt.Printf("Walk failed %v", err)
	}
}
