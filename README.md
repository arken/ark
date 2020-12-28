# ait

The Arken Import Tool allows anyone to index and upload data to an Arken
cluster.

[![Go Report Card](https://goreportcard.com/badge/github.com/arkenproject/ait)](https://goreportcard.com/report/github.com/arkenproject/ait)

## What is the Arken Import Tool?

You can think of the Arken Import Tool or AIT as a git like upload client for
Arken that indexes, generates, and submits Keyset additions as pull requests.

## Installation

1. Go to AIT Releases
2. Copy the link to your corresponding OS and Architecture.
3. Run sudo curl -L "PATH-TO-RELEASE" -o /usr/local/bin/ait
4. sudo chmod a+x /usr/local/bin/ait
5. (Optional) sudo ln -s /usr/local/bin/ait /usr/bin/ait

## Usage

### Commands

| Command             |  Alias  | Description                                                                |
| ------------------- | ------- | -------------------------------------------------------------------------- |
| `help`              | `?`     | Get help with a specific subcommand.                                       |
| `stage`             | `st`    | Stage files or directories for submission.                            |
| `init`              | `i`     | Initialize a dataset's local configuration.                                |
| `unstage`           | `un`    | Remove files or directories from AIT's staged files.                       |
| `remote`            | `r`     | Allows the use of aliases for commonly used URLs.                          |
| `status`            | `s`     | View what files are currently staged for submission.                       |
| `submit`            | `sm`    | Submit your Keyset to a git keyset repository.                             |
| `upload`            | `up`    | After Submitting Your Files upload Them to the Arken Cluster.              |
| `pull`              | `pl`    | Pull one or many files from the Arken Cluster.                             |
| `update`            | `upd`   | Have AIT update its own binary.                                            |

### Tutorial

#### Initializing a KeySet

Go to the location of your data and run.

```bash
ait init
```

#### Stage Data to Your KeySet Submission

Still within the location of your data add specific files or folders.

```bash
ait stage <LOCATION>
```

##### ex.

Stage the example.csv file into your Arken Submission.

```bash
ait stage example.csv
```

or to stage everything within the folder containing your data.

```bash
ait stage .
```

#### Submit Your Data to the KeySet

This will index the added data, generate a keyset file, and either add that file
to the remote git repository or generate a pull request if you don't have access
to the main repository.

```bash
ait submit <KEYSET-LOCATION>
```

##### ex.

Submit your data to the official
curated [Core Arken Keyset](https://github.com/arkenproject/core-keyset).

```bash
ait submit https://github.com/arkenproject/core-keyset
```
