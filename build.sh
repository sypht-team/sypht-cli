  
#!/bin/bash

set -e

rm -rf build
mkdir build

go build *.go

GOOS=windows GOARCH=amd64 go build -o sypht-cli.exe *.go
