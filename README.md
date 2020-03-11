# sypht-cli
[![Build Status](https://travis-ci.com/sypht-team/sypht-cli.svg?branch=master)](https://travis-ci.com/sypht-team/sypht-cli)

Sypht-cli is a tool for uploading documents to Sypht's API.

# About Sypht
[Sypht](https://sypht.com) is a SaaS [API]((https://docs.sypht.com/)) which extracts key fields from documents. For 
example, you can upload an image or pdf of a bill or invoice and extract the amount due, due date, invoice number 
and biller information. 

# Getting started
To get started you'll need API credentials, i.e. a `client_id` and `client_secret`, which can be obtained by registering
for an [account](https://www.sypht.com/signup/developer).
For enterprise users please contact us for more details.

Navigate to our github [release page](https://github.com/sypht-team/sypht-cli/releases) to download Sypht-cli that suits your operating system.

Extract the assets.zip file and edit config.json to use your `client_id` and `client_secret`:
```Bash
{
    "ClientID" : "82uEDuQ...",
    "ClientSecret" : "yIO03P-..."
}
```
Make sure your `config.json` file lives in the same directory with `sypht-cli` or `sypht-cli.exe`. 
# Usage
```Bash
sypht-cli command [command options] [arguments...]
```
Available command
```Bash
scan     sypht-cli scan [OPTIONS] [directory]

DESCRIPTION:
   Scan and upload all documents in a directory to Sypht API.

OPTIONS:
   --rate-limit value  Number of files to upload per second (default: 1)
   --recursive, -R     Recursively scan files in subdirectories (default: false)

watch    sypht-cli watch [OPTIONS] [directory]

DESCRIPTION:
   Watch and upload all newly added documents in a directory to Sypht API.

OPTIONS:
   --rate-limit value  Number of files to upload per second (default: 1)
   --recursive, -R     Recursively watch files in subdirectories (default: false)
```
If [directory] is omitted, current directory is used by default.

## Running on Windows Operating System
Run Command Prompt as administrator, navigate to the folder where `sypht-cli.exe` and `config.json` live.
```cmd
start sypht-cli.exe scan --rate-limit 2 -recursive C:\Users\Administrator\Desktop\
```
This command will do a recursive scan of directory `C:\Users\Administrator\Desktop\` and it's subdirectories with a rate of 2 documents per second, and upload all valid documents to Sypht API.


```cmd
start sypht-cli.exe watch C:\Users\Administrator\Desktop\
```
This command will watch non-recursively on directory `C:\Users\Administrator\Desktop\` only, with a rate of 1 documents per second. It will detect any newly added files in directory `C:\Users\Administrator\Desktop\`, and upload them to Sypht API.

Note : Watch does NOT scan the directory.

Upload result can be found in `[filename].json` and `sypht.csv`.

Supported document types:  `.pdf .png .jpg .jpeg .tiff .tif .gif`


# License
The software in this repository is available as open source under the terms of the [MIT License](https://github.com/sypht-team/sypht-cli/blob/master/LICENSE).


