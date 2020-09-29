package types

import "sync"

type ThreadSafeSet struct {
    internal map[string]struct{}
    sync.RWMutex
}

func NewThreadSafeSet() *ThreadSafeSet {
    return  &ThreadSafeSet{
        internal: make(map[string]struct{}),
    }
}

func (rss *ThreadSafeSet) Contains(str string) bool {
    rss.RLock()
    _, ok := rss.internal[str]
    rss.RUnlock()
    return ok
}

func (rss *ThreadSafeSet) Add(str string) {
    rss.Lock()
    rss.internal[str] = struct{}{}
    rss.Unlock()
}

func (rss *ThreadSafeSet) Delete(str string) {
    rss.Lock()
    delete(rss.internal, str)
    rss.Unlock()
}

func (rss *ThreadSafeSet) Size() int {
    rss.RLock()
    res := len(rss.internal)
    rss.RUnlock()
    return res
}
