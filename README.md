# ait
The Arken Import Tool allows anyone to index and upload data to an Arken cluster. 

[![Go Report Card](https://goreportcard.com/badge/github.com/arkenproject/ait)](https://goreportcard.com/report/github.com/arkenproject/ait)

## What is the Arken Import Tool?
You can think of the Arken Import Tool or AIT as a git like upload client for Arken that indexes, generates, and submits Keyset additions as pull requests.

## Installation
1. Install Go(lang)
2. go get github.com/arkenproject/ait

## Usage
### Commands

| Command           | Alias | Discription                                                                |
| ----------------- | ----- | -------------------------------------------------------------------------- |
| `help`              | `?`     | Get help with a specific subcommand.                                       |
| `add`               | `a`     | Add a file or directory to AIT's tracked files.                            |
| `init`              | `i`     | Initialize a dataset's local configuration.                                |
| `remove`            | `rm`    | Remove a file or directory from AIT's tracked files.                       |
| `remote`            | `r`     | Allows the use of aliases for commonly used URLs.                          |
| `status`            | `s`     | View what files are currently staged for submission.                       |
| `submit`            | `sm`    | Submit your Keyset to a git keyset repository.                             |
| `upload`            | `up`    | After Submitting Your Files upload Them to the Arken Cluster.              |

### Tutorial
#### Initializing a KeySet
Go to the location of your data and run.
```bash
ait init
```

#### Stage Data to Your KeySet Submission
Still within the location of your data add specific files or folders.
```bash
ait add <LOCATION>
```

##### ex.
Stage the example.csv file into your Arken Submission.
```bash
ait add example.csv
```
or to stage everything within the folder containing your data.
```bash
ait add .
```

#### Submit Your Data to the KeySet
This will index the added data, generate a keyset file, and either add that file to the remote git repository or generate a pull request if you don't have access to the main repository.
```bash
ait submit <KEYSET-LOCATION>
```
##### ex.
Submit your data to the official curated [Core Arken Keyset](https://github.com/arkenproject/core-keyset).
```bash
ait submit https://github.com/arkenproject/core-keyset
```
