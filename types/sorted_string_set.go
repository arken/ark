package types

import (
	"github.com/google/btree"
)

// SortedStringSet is a simple container around a B-tree. It has the
// functionality and performance of a tree-based set, because it is one.
type SortedStringSet struct {
	internal *btree.BTree
}

// NewSortedStringSet returns a pointer to a new SortedStringSet object
func NewSortedStringSet() *SortedStringSet {
	return  &SortedStringSet{
		internal: btree.New(2),
	}
}

// Contains returns true if the set contains str, false otherwise
func (set *SortedStringSet) Contains(str string) bool {
	return set.internal.Has(String{str})
}

// Add adds the given string to the set
func (set *SortedStringSet) Add(str string) {
	set.internal.ReplaceOrInsert(String{str})
}

// Delete deletes the given string from the set
func (set *SortedStringSet) Delete(str string) {
	set.internal.Delete(String{str})
}

// Size returns the amount of items in the set.
func (set *SortedStringSet) Size() int {
	return set.internal.Len()
}

// Performs the given function on each element of the set
func (set *SortedStringSet) ForEach(f func(s string) error) error {
	var err error = nil
	set.internal.Ascend(func(item btree.Item) bool {
			str := item.(String)
			err = f(str.s)
			if err != nil {
				//stops the iteration
				return false
			}
			return true
			// ^ This makes it go through the whole tree
	})
	return err
}
