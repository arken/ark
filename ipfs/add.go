package ipfs

import (
	"os"

	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/options"
)

// Add imports a file to IPFS and returns the file identifier to ait.
func Add(path string) (cid string, err error) {
	file, err := getUnixfsFile(path)
	if err != nil {
		return nil, err
	}
	return ipfs.Unixfs().Add(ctx, node, func(input *options.UnixfsAddSettings) {
		input.FsCache = false
		input.Pin = true
	})
}

// getUnixfsFile opens a file to read for IPFS import.
func getUnixfsFile(path string) (files.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return nil, err
	}

	f, err := files.NewReaderPathFile(path, file, st)
	if err != nil {
		return nil, err
	}

	return f, nil
}
