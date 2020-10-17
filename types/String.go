package types

import (
	"github.com/google/btree"
	"strings"
)

// String is for use in the sorted string set. It implements the btree.Item
// interface, which just contains the Less method.
type String struct {
	s string
}


func (str String) Less(than btree.Item) bool {
	other := than.(String)
	return strings.Compare(str.s, other.s) < 0
}
