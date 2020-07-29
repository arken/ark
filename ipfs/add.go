package ipfs

import (
	"os"

	files "github.com/ipfs/go-ipfs-files"
)

// Add imports a file to IPFS and returns the file identifier to ait.
func Add(path string) (cid string, err error) {
	file, err := getUnixfsNode(path)
	if err != nil {
		return cid, err
	}
	output, err := ipfs.Unixfs().Add(ctx, file)
	if err != nil {
		return cid, err
	}
	cid = output.Cid().String()
	return cid, nil
}

func getUnixfsNode(path string) (files.Node, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := files.NewSerialFile(path, false, st)
	if err != nil {
		return nil, err
	}

	return f, nil
}
