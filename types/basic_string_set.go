package types

// BasicStringSet is a simple container around a map[string]struct{}, whose values
// have 0 width, so it has the functionality and performance of a hash-based set.
type BasicStringSet struct {
    internal map[string]struct{}
}

// NewBasicStringSet returns a pointer to a new BasicStringSet object
func NewBasicStringSet() *BasicStringSet {
    return  &BasicStringSet{
        internal: make(map[string]struct{}),
    }
}

// Contains returns true if the set contains str, false otherwise
func (set *BasicStringSet) Contains(str string) bool {
    _, ok := set.internal[str]
    return ok
}

// Add adds the given string to the set
func (set *BasicStringSet) Add(str string) {
    set.internal[str] = struct{}{}
}

// Delete deletes the given string from the set
func (set *BasicStringSet) Delete(str string) {
    delete(set.internal, str)
}

// Size returns the amount of items in the set.
func (set *BasicStringSet) Size() int {
    return len(set.internal)
}

// ForEach performs the given function on each element of the set
func (set *BasicStringSet) ForEach(f func(s string) error) error {
    var err error = nil
    for str := range set.internal {
        err = f(str)
        if err != nil {
            break
        }
    }
    return err
}
