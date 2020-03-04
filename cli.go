package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type config struct {
	ClientID     string `json:"ClientID"`
	ClientSecret string `json:"ClientSecret"`
}

type uploadResult struct {
	fileID     string
	path       string
	uploadedAt string
	status     string
}

func getConfig(file string) (config config) {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errChan := make(chan error, 1)
	go func() {
		// Close the paths channel after Walk returns.
		defer close(paths)
		// No select needed for this send, since errChan is buffered.
		errChan <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			select {
			case paths <- path:
			case <-done:
				return errors.New("walk canceled")
			}
			return nil
		})
	}()
	return paths, errChan
}

func processFile(done <-chan struct{}, paths <-chan string, c chan<- uploadResult) {
	for path := range paths {
		// upload file here
		select {
		case c <- uploadResult{
			fileID:     path,
			path:       path,
			uploadedAt: time.Now().String(),
			status:     "good",
		}:
		case <-done:
			return
		}
	}
}

func watch(path string, workflow string, rateLimit int32, recursive bool) error {
	if recursive {

	}
	return nil
}

func main() {
	currentDirectory, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	getConfig(filepath.Join(currentDirectory, "config.json"))

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
		log.Fatalf("Walk failed %v", err)
	}

	// client, err := syphtClient.NewSyphtClient(fmt.Sprintf("%s:%s", config.ClientID, config.ClientSecret), nil)
	// if err != nil {
	// 	log.Fatalf("Unable to start Sypht client , %v", err)
	// }
	// uploaded, err := client.Upload("./sample_invoice.pdf", []string{
	// 	sypht.Invoice,
	// 	sypht.Document,
	// })
	// if err != nil {
	// 	log.Printf("Unable to start Sypht client , %v", err)
	// } else {
	// 	log.Print(uploaded)
	// }
}
