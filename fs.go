package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

const UploadStatusFAILED = "FAILED"

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

func walkDirs(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	dirs := make(chan string)
	errChan := make(chan error, 1)
	go func() {
		// Close the dirs channel after Walk returns.
		defer close(dirs)
		// No select needed for this send, since errChan is buffered.
		errChan <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			select {
			case dirs <- path:
			case <-done:
				return errors.New("walk canceled")
			}
			return nil
		})
	}()
	return dirs, errChan
}

func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errChan := make(chan error, 1)
	if cliFlags.recursive {
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
	} else {
		go func() {
			defer close(paths)
			// Close the paths channel after reading dirs.
			files, err := ioutil.ReadDir(root)
			if err != nil {
				errChan <- err
			}
			for _, file := range files {
				select {
				default:
					if file.Mode().IsRegular() && !file.IsDir() {
						paths <- filepath.Join(root, file.Name())
					}
				case <-done:
					errChan <- errors.New("Scan canceled")
				}
			}
			errChan <- nil
		}()
	}

	return paths, errChan
}

func processFile(done <-chan struct{}, paths <-chan string, c chan<- uploadResult, tickerChan <-chan time.Time) {
	select {
	case _ = <-tickerChan:
		for path := range paths {
			var result *uploadResult
			ok := validateFile(path)
			if !ok {
				continue
			}
			resp, err := uploadFileWithRetry(path, 10)
			if err != nil {
				fmt.Printf("Error uploading file path %s, msg %v", path, err)
				result = &uploadResult{
					Path:   path,
					Status: UploadStatusFAILED,
					Error:  err.Error(),
				}
			} else {
				var status string
				if resp.Status != "" {
					status = resp.Status
				} else {
					status = UploadStatusFAILED
				}
				result = &uploadResult{
					FileID:     resp.FileID,
					Path:       path,
					UploadedAt: resp.UploadedAt,
					Status:     status,
					Error:      resp.Code,
				}
			}

			select {
			case c <- *result:
				base, _ := extractFileInfo(path)
				f, err := os.Create(fmt.Sprintf("%s.json", base))
				if err != nil {
					fmt.Printf("Error creating result file path %s, msg %v", path, err)
				}
				if f != nil {
					resultJSON, _ := json.Marshal(result)
					f.Write(resultJSON)
					f.Close()
				}

				data, err := readFileWithRetry(path, 10)
				if err != nil {
					fmt.Printf("Error reading file after retry path %s, msg %v", path, err)
				}
				checksum := md5.Sum(data)

				metaFileLock.Lock()
				var errStr string
				if err != nil {
					errStr = err.Error()
				}
				csvWriter.Write([]string{result.FileID, result.Path, result.Status, result.UploadedAt, errStr, hex.EncodeToString(checksum[:])})
				csvWriter.Flush()
				metaFileLock.Unlock()
			case <-done:
				return
			}
		}
	}
}

func readFileWithRetry(path string, times int) (data []byte, err error) {
	data, err = ioutil.ReadFile(path)
	for err != nil && times > 0 {
		time.Sleep(time.Second)
		times--
		data, err = ioutil.ReadFile(path)
	}
	return
}

func uploadFileWithRetry(path string, times int) (resp sypht.UploadResponse, err error) {
	resp, err = client.Upload(path, []string{}, cliFlags.workflowID)
	for err != nil && times > 0 {
		time.Sleep(time.Second)
		times--
		resp, err = client.Upload(path, []string{}, cliFlags.workflowID)
	}
	return
}
