  
#!/bin/bash

set -e

rm -rf build
mkdir build

go build -o sypht-cli *.go

GOOS=windows GOARCH=amd64 go build -o sypht-cli.exe *.go

cp config.json build/
mv sypht-cli* build/

zip -r assets.zip build/
mv assets.zip build/