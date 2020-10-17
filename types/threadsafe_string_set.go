package types

import "sync"

// ThreadSafeStringSet is like a BasicStringSet, except that it uses an RWMutex
// to enforce thread safety.
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

// Contains returns true if the set contains str, false otherwise.
// No lock.
func (set *ThreadSafeStringSet) Contains(str string) bool {
	_, ok := set.internal[str]
	return ok
}

// Add adds the given string to the set.
// Write lock.
func (set *ThreadSafeStringSet) Add(str string) {
	set.Lock()
	set.internal[str] = struct{}{}
	set.Unlock()
}

// Delete deletes the given string from the set
// Write lock.
func (set *ThreadSafeStringSet) Delete(str string) {
	set.Lock()
	delete(set.internal, str)
	set.Unlock()
}

// Size returns the amount of items in the set.
// No lock.
func (set *ThreadSafeStringSet) Size() int {
	return len(set.internal)
}

// Performs the given function on each element of the set.
// Write lock.
func (set *ThreadSafeStringSet) ForEach(f func(s string) error) error {
	var err error = nil
	set.Lock()
	for str := range set.internal {
		err = f(str)
		if err != nil {
			break
		}
	}
	set.Unlock()
	return err
}

// Underlying returns the underlying map. This is only meant for ranging purposes.
// No lock (So don't call this in a multithreaded context).
func (set *ThreadSafeStringSet) Underlying() map[string]struct{} {
	return set.internal
}

