# Zipspy
Zipspy is a CLI tool to extract files from zip archives in S3 without needing to download the entire archive

<!-- TOC depthFrom:1 depthTo:3 withLinks:1 updateOnSave:1 orderedList:0 -->

- [Zipspy](#zipspy)
    - [Purpose](#purpose)
    - [Installation](#installation)
    - [Instructions](#instructions)
    - [Examples](#examples)
        - [List](#list)
        - [Extract](#extract)
    

<!-- /TOC -->



## Purpose

Zipspy allows you to search for and download files from remote archives quicker. For example, imagine you have a 10MB file (compressed) within a 10GB zip archive in AWS S3. Instead of downloading the entire 10GB file, extracting your file, and deleting the excess, Zipspy downloads _only_ the 10MB compressed file you care about (plus 1-65K for the central directory). 

## Installation

To install the Zipspy, run the following command:

```
$ go install github.com/alec-rabold/zipspy
```

You may check that it's installed correctly by running:

```
$ zipspy -v
zipspy version db00fe2
```

## Instructions

Zipspy currently supports reading from two storage locations:
- AWS S3 Bucket (`s3://`)
- Local File on Disk (`file://`) [_note_: mainly for development]

The underlying providers for each are determined by the protocol specified in the global, required flag `--location`.

For example, an S3 location may look like `"s3://my-bucket/archive.zip"` while a local file location would look like `file://path/to/archive.zip`.

For S3, all AWS configuration will be read from your environment through the [shared config functionality](https://docs.aws.amazon.com/sdkref/latest/guide/creds-config-files.html). 

To see all available commands, simply type `zipspy`:
```
$ zipspy 
                       
 ____  __  ____  ____  ____  _  _ 
(__  )(  )(  _ \/ ___)(  _ \( \/ )
 / _/  )(  ) __/\___ \ ) __/ )  / 
(____)(__)(__)  (____/(__)  (__/    

Zipspy allows you interact with ZIP archives stored in remote locations without
requiring a local copy. For example, you can list the filenames in an S3 ZIP archive, 
download a subset of files, search and retrieve files with regular expressions, and more!

Usage:
  zipspy [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  extract     Extract one or more files from the zip archive.
  help        Help about any command
  list        List all file names from a zip archive.

Flags:
      --development        whether or not to use development settings
  -h, --help               help for zipspy
      --location string    (required) protocol and address of your ZIP archive ("file://archive.zip", "s3://<bucket_name>/archive.zip")
      --verbosity string   global log level (trace, debug, info, warn, error, fatal, panic) (default "warning")
  -v, --version            version for zipspy

Use "zipspy [command] --help" for more information about a command.
```

As indicated, you may also use `zipspy [command] --help` for more information about a command:
```
$ zipspy list --help

Prints out the names of all files contained within a zip archive.

Usage:
  zipspy list [--include-directory-names] [flags]

Flags:
  -h, --help                      help for list
      --include-directory-names   (optional) include the leaf names of directories
      --no-newlines               (optional) omit the newlines appended to file names
  -o, --out string                (optional) name of a file to write output to
      --separator string          (optional) separator when combining the output of multiple file names

Global Flags:
      --development       whether or not to use development settings
      --location string   (required) protocol and address of your ZIP archive ("file://archive.zip", "s3://<bucket_name>/archive.zip")
```

## Examples
Suppose you have a zip file named `archive.zip` at the top level of an AWS S3 bucket called `my-bucket`. 

The contents of `archive.zip` are as follows:
```
test/
├── important.txt
├── path
│   ├── bin
│   │   └── program
│   └── to
│       ├── file.txt
│       └── movie.mp4
└── test.txt
```

### List

To list all the file names, run the following command:
```
$ zipspy list --location "s3://my-bucket/archive.zip"
archive/important.txt
archive/path/bin/program
archive/path/to/file.txt
archive/path/to/movie.mp4
archive/test.txt
```

You may choose to include bare directory names by using the `--include-directory-names` flag:
```
$ zipspy list --location "s3://my-bucket/archive.zip" --include-directory-names
test/
test/important.txt
test/path/
test/path/bin/
test/path/bin/program
test/path/to/
test/path/to/file.txt
test/path/to/movie.mp4
test/test.txt
```

### Extract

To extract a particular file, use the `extract` command:
```Shell
$ zipspy extract --location "s3://my-bucket/archive.zip" -f "archive/important.txt"
Contents of important document.
```

You extract multiple files simultaneously, separated by a newline by default:
```Shell
$ zipspy extract --location "s3://my-bucket/archive.zip" -f "archive/important.txt" -f "archive/path/to/file.txt"
Contents of important document.
Notes from file.
```

If the number of files provided does not match the number of files found, a warning will be output:
```Shell
$ zipspy extract --location "s3://my-bucket/archive.zip" -f "archive/important.txt" -f "archive/DNE.jpg"
WARN[0000] number of input files does not match number of found files (input: 2) (found: 1) 
Contents of important document.
```
You may silence these warning with the `--silence-warnings` flag.

To use a custom separator, use the `--separator` flag:
```Shell
$ zipspy extract --location "s3://my-bucket/archive.zip" -f "archive/important.txt" -f "archive/path/to/file.txt" --separator "---"
Contents of important document.
---
Notes from file.
```

## Development

Zipspy uses a plugin-based architecture. A plugin must simply satisfy the `zipspy.Reader` interface:
```Go
type Reader interface {
	io.ReaderAt
	Size() (int64, error)
}
```

For remote locations, it's preferable to use [HTTP Range Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests) where possible.

While this will likely produce a greater number of requests, the target consumers for zipspy will benefit from substantially greater speed and lower network consumption. 

---

### Thanks for taking a look! Feature requests and contributions are welcomed.

---
