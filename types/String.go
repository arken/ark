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

// Less is basically a wrapper around strings.Compare for the btree.Item
// interface. This method has an unsafe cast from btree.Item to String, and
// if this cast turns out not to be safe, there will be a panic. This is expected
// behavior since in AIT's context, a btree of Strings should only have strings.
func (str String) Less(than btree.Item) bool {
	other := than.(String)
	return strings.Compare(str.s, other.s) < 0
}
