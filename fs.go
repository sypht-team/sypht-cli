package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sypht-team/sypht-golang-client"
)

var metaFileLock sync.Mutex
var csvWriter *csv.Writer

type uploadResult struct {
	FileID     string `json:"fileId"`
	Path       string `json:"path"`
	UploadedAt string `json:"uploadedAt"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

func extractFileInfo(path string) (base, ext string) {
	ext = strings.ToLower(filepath.Ext(path))
	base = path[0 : len(path)-len(ext)]
	return
}

func validateFile(path string) (ok bool) {
	exts := map[string]interface{}{".pdf": nil, ".png": nil, ".jpg": nil, ".jpeg": nil, ".tiff": nil, ".tif": nil, ".gif": nil}
	base, fileExt := extractFileInfo(path)

	_, ok = exts[fileExt]
	if !ok {
		return
	}
	if _, err := os.Stat(base + ".json"); os.IsNotExist(err) {
		ok = true
	} else {
		ok = false
	}
	return
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

func processFile(done <-chan struct{}, paths <-chan string, c chan<- uploadResult, rate int) {
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	for path := range paths {
		result := &uploadResult{}
		ok := validateFile(path)
		if !ok {
			continue
		}
		resp, _ := uploadFile(path)
		if _, ok = resp["message"]; ok {
			if _, ok = resp["code"]; ok {
				result = &uploadResult{
					Path:   path,
					Status: resp["code"].(string),
					Error:  resp["message"].(string),
				}
			}
		} else {
			result = &uploadResult{
				FileID:     resp["fileId"].(string),
				Path:       path,
				UploadedAt: resp["uploadedAt"].(string),
				Status:     resp["status"].(string),
			}
		}

		select {
		case _ = <-ticker.C:
			c <- *result
			base, _ := extractFileInfo(path)
			f, err := os.Create(fmt.Sprintf("%s.json", base))
			if err != nil {
				fmt.Printf("Error creating result file path %s, msg %v", path, err)
			}
			defer f.Close()

			resultJSON, _ := json.Marshal(result)
			f.Write(resultJSON)

			data, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file path %s, msg %v", path, err)
			}
			checksum := md5.Sum(data)

			metaFileLock.Lock()
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			csvWriter.Write([]string{result.FileID, result.Path, result.Status, result.UploadedAt, errStr, hex.EncodeToString(checksum[:])})
			metaFileLock.Unlock()
		case <-done:
			ticker.Stop()
			return
		}
	}
}

func uploadFile(path string) (resp map[string]interface{}, err error) {
	resp, err = client.Upload(path, []string{
		sypht.Invoice,
		sypht.Document,
	}, cliFlags.workflowID)
	if err != nil {
		log.Printf("Error uploading file %s , %v", path, err)
	}
	return
}
