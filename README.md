# ark
A Command Line Client for Arken Clusters

[![Go Report Card](https://goreportcard.com/badge/github.com/arken/ark)](https://goreportcard.com/report/github.com/arken/ark)

## What is Ark?

Ark is a command line client for Arken that indexes, generates, and submits manifest additions as pull requests.
Ark can also directly download collections of files from the nodes within an Arken cluster.

## Installation

1. Go to Ark Releases (over there -->)
2. Copy the link to your corresponding OS and Architecture.
3. Run `sudo curl -L "PATH-TO-RELEASE" -o /usr/local/bin/ark`
4. Run `sudo chmod a+x /usr/local/bin/ark`
5. (Optional) Run `sudo ln -s /usr/local/bin/ark /usr/bin/ark`

## Usage

### Commands

| Command             |  Alias  | Description                                                                |
| ------------------- | ------- | -------------------------------------------------------------------------- |
| `help`              | `?`     | Get help with a specific subcommand.                                       |
| `stage`             | `st`    | Stage files or directories for submission.                                 |
| `init`              | `i`     | Initialize a dataset's local configuration.                                |
| `unstage`           | `un`    | Remove files or directories from Ark's staged files.                       |
| `remote`            | `r`     | Allows the use of aliases for commonly used URLs.                          |
| `status`            | `s`     | View what files are currently staged for submission.                       |
| `submit`            | `sm`    | Submit your manifest to a git manifest repository.                         |
| `upload`            | `up`    | After Submitting Your Files upload Them to the Arken Cluster.              |
| `pull`              | `pl`    | Pull one or many files from the Arken Cluster.                             |
| `update`            | `upd`   | Have AIT update its own binary.                                            |

### Tutorial

#### Initializing a manifest

Go to the location of your data and run. (If you're running MacOS or Linux you can navigate to the folder containing your data
in your file browser/finder and by right clicking on the folder open a terminal at that location.)

```bash
ark init
```

#### Stage Data to Your manifest Submission

Still within the location of your data add specific files or folders.

```bash
ark stage <LOCATION>
```

##### ex.

Stage the example.csv file into your Arken Submission.

```bash
ark stage example.csv
```

or to stage everything within the folder containing your data.

```bash
ark stage .
```

#### Submit Your Data to the manifest

This will index the added data, generate a manifest file, and either add that file
to the remote git repository or generate a pull request if you don't have access
to the main repository.

```bash
ark submit <manifest-LOCATION>
```

##### ex.

Submit your data to the official
curated [Core Arken manifest](https://github.com/arken/core-manifest).

```bash
ark submit https://github.com/arken/core-manifest
```

#### Uploading Your Data After Your Submission Has Been Accepted

After your submission is accepted you'll receive an email notifying you the Pull Request
has been merged into the manifest. At this point you can finally run ark upload from the directory with
your data in it to upload the data to the cluster. 
```bash
ark upload
```

*Note:* If you attempt to run `ark upload` before your submission is accepted your data will not begin syncing with the cluster.

## License

Copyright 2019-2021 Alec Scott & Arken Project <team@arken.io>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
