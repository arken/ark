package types

import "sync"

// ThreadSafeStringSet is a simple container around a map[string]struct{}, whose values
// have 0 width, so it has the functionality and performance of a hash-based set.
type ThreadSafeStringSet struct {
	internal map[string]struct{}
	sync.RWMutex
}

// NewThreadSafeStringSet returns a pointer to a new ThreadSafeStringSet object
func NewThreadSafeStringSet() *ThreadSafeStringSet {
	return  &ThreadSafeStringSet{
		internal: make(map[string]struct{}),
	}
}

// Contains returns true if the set contains str, false otherwise
func (set *ThreadSafeStringSet) Contains(str string) bool {
	_, ok := set.internal[str]
	return ok
}

// Add adds the given string to the set
func (set *ThreadSafeStringSet) Add(str string) {
	set.Lock()
	set.internal[str] = struct{}{}
	set.Unlock()
}

// Delete deletes the given string from the set
func (set *ThreadSafeStringSet) Delete(str string) {
	set.Lock()
	delete(set.internal, str)
	set.Unlock()
}

// Size returns the amount of items in the set.
func (set *ThreadSafeStringSet) Size() int {
	return len(set.internal)
}

// Performs the given function on each element of the set
func (set *ThreadSafeStringSet) ForEach(f func(s string)) {
	set.Lock()
	set.RLock()
	for str := range set.internal {
		f(str)
	}
	set.Unlock()
	set.RUnlock()
}

func (set *ThreadSafeStringSet) Underlying() map[string]struct{} {
	return set.internal
}

