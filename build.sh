  
#!/bin/bash
set -e

package_name="sypht-cli"

platforms=("windows/amd64" "windows/386" "darwin/amd64")

rm -rf build
mkdir build

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$GOOS'-'$GOARCH
    zip_name=$output_name'.zip'
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
    zip $zip_name $output_name config.json
    mv $zip_name build/
done

