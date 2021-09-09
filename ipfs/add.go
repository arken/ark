package ipfs

import (
	"os"

	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/options"
)

// Add imports a file to IPFS and returns the file identifier to Ark.
func (n *Node) Add(path string, onlyHash bool) (cid string, err error) {
	file, err := getUnixfsNode(path)
	if err != nil {
		if file != nil {
			file.Close()
		}
		return cid, err
	}
	output, err := n.api.Unixfs().Add(n.ctx, file, func(input *options.UnixfsAddSettings) error {
		input.Pin = true
		input.NoCopy = true
		input.CidVersion = 1
		input.OnlyHash = onlyHash
		return nil
	})
	if err != nil {
		return cid, err
	}
	cid = output.Cid().String()
	file.Close()
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
