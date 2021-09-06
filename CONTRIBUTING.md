# How to Contribute to the AIT Application

If you're looking for a place to start, go to the issues tab after reading this
document and help add a new feature or fix a bug. If there aren't any issues
we're always looking for people to test the commandline tool or find important
open source data to upload to the
[Arken Project Core Repository](https://github.com/arken/core-manifest).

## What's a manifest anyway?

A manifest is like a ship's manifest. It contains a list of all the file names &
IPFS identifiers of the files to be added to an Arken cluster WITHOUT actually
containing any of the files' raw data.

A manifest repository is made up of manifest (or keyset) files which look like this,

##### node.ks

``` plain
QmYyLws3LmM85EfgNrEgGENoG8LPcKnZHR87A7BbgFqKsf NODE_VOL_01.pdf
QmRRhKLebvXztobrhJifNLVgQJA4TDfv1tQV9RsVoLnsS4 NODE_VOL_02.pdf
```

While using AIT, a user should **never** directly deal with a manifest file or 
repository. All files should be generated automatically and handled in the
background of a publish command run. Researchers/users should only care about
their data and not have to also deal with manifest files themselves.

## Project Structure

```plain
/apis    --> api library for API calls to external services.
/cli     --> cli library that defines the application's publicly available commands.
/config  --> config library that defines defaults and engine for reading and generating ark's global configuration.
/display --> display library for showing users a text editor when editing their applications.
/ipfs    --> ipfs library providing functions for creating a node, adding files to the ipfs network, etc...
/manifests --> manifests library providing functions for adding, removing, pulling, and pushing manifest repositories.
/types   --> types library containing non-trivially small type declarations for use throughout the app.
```

## Project Conventions

- Code should be formatted using Go standard conventions. Use `go fmt -s` for linting.
- Minimize the number of unnecessary public functions.
- Write tests for all added functions to test expected functionality.
- Start function comments with the name of the function.
