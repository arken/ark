# How to Contibute to the AIT Application
If you're looking for a place to start, go to the issues tab after reading this document and help add a new feature or fix a bug. If there aren't any issues
we're always looking for people to test the commandline tool or find important open source data to upload to the 
[Arken Project Core Repository](https://github.com/arkenproject/core-keyset).

## What's a Keyset anyway?
A keyset is like a ships manifest. It contains a list of all the file names & IPFS identifiers of the files to be added to an Arken cluster
WITHOUT actually containing any of the raw data.

A Keyset Repository is made up of Keyset files which look like this,
##### node.ks
``` plain
NODE_VOL_01.pdf QmYyLws3LmM85EfgNrEgGENoG8LPcKnZHR87A7BbgFqKsf
NODE_VOL_02.pdf QmRRhKLebvXztobrhJifNLVgQJA4TDfv1tQV9RsVoLnsS4
```

Using AIT a user should **never** directly deal with a Keyset file or repository. All files should be generated automatically and handled in the background
of a publish command run. Researchers/Users should only care about their data and not have to also deal with keyset files.

## Project Structure
```plain
/cli     --> cli library that defines the application's publicly availble commands.
/config  --> config library that defines defaults and engine for reading and generating ait's global configuration.
/ipfs    --> ipfs library providing functions for creating a node, adding files to the ipfs network, etc...
/keysets --> keysets library providing functions for adding, removing, pulling, and pushing keyset repositories. 
```

## Project Convensions
- Code should be formatted using Go standard conventions. Use `go fmt -s` for linting.
- Minimize the number of public functions that do not need to be public.
- Write tests for all added functions to test expected functionality.
- Start function comments with the name of the function.
