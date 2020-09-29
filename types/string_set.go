package types

// StringSet is a simple container around a map[string]struct{}, whose values
// have 0 width, so it has the functionality and performance of a hash-based set.
type StringSet struct {
    internal map[string]struct{}
}

// NewStringSet returns a pointer to a new StringSet object
func NewStringSet() *StringSet {
    return  &StringSet{
        internal: make(map[string]struct{}),
    }
}

// Contains returns true if the set contains str, false otherwise
func (set *StringSet) Contains(str string) bool {
    _, ok := set.internal[str]
    return ok
}

// Add adds the given string to the set
func (set *StringSet) Add(str string) {
    set.internal[str] = struct{}{}
}

// Delete deletes the given string from the set
func (set *StringSet) Delete(str string) {
    delete(set.internal, str)
}

// Size returns the amount of items in the set.
func (set *StringSet) Size() int {
    return len(set.internal)
}

// Performs the given function on each element of the set
func (set *StringSet) ForEach(f func(s string)) {
    for str := range set.internal {
        f(str)
    }
}
