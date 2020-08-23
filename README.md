# ait
The Arken Import Tool allows anyone to index and upload data to an Arken cluster. 

## What is the Arken Import Tool?
You can think of the Arken Import Tool or AIT as a git like upload client for Arken that indexes, generates, and submits Keyset additions as pull requests.

## Installation
1. Install Go(lang) 1.14
2. go get github.com/arkenproject/ait

## Usage
##### Initializing a KeySet
Go to the location of your data and run.
```bash
ait init
```

##### Configure the Remote Keysets
```bash
ait add-remote https://github.com/arkenproject/core-keyset
```

##### Adding Data to Your KeySet Submission
```bash
ait add <LOCATION>
```

ex.
```bash
ait add .
```

##### Submit Your Data to the KeySet
This will index the added data, generate a keyset file, and either add that file to the remote git repository or generate a pull request if you don't have access to the main repository.
```bash
ait submit
```
