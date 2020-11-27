package ipfs

import (
	"errors"

	files "github.com/ipfs/go-ipfs-files"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"
)

// Pull grabs the requested file from the IPFS network and returns
// it as a string.
func Pull(cid string) (output files.Node, err error) {
	path := icorepath.New("/ipfs/" + cid)

	output, err = ipfs.Unixfs().Get(ctx, path)
	if err != nil {
		return output, errors.New("Could not get file with CID: " + cid)
	}

	return output, nil
}
