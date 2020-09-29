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

func (tss *ThreadSafeSet) Contains(str string) bool {
    tss.RLock()
    _, ok := tss.internal[str]
    tss.RUnlock()
    return ok
}

func (tss *ThreadSafeSet) Add(str string) {
    tss.Lock()
    tss.internal[str] = struct{}{}
    tss.Unlock()
}

func (tss *ThreadSafeSet) Delete(str string) {
    tss.Lock()
    delete(tss.internal, str)
    tss.Unlock()
}

func (tss *ThreadSafeSet) Size() int {
    tss.RLock()
    res := len(tss.internal)
    tss.RUnlock()
    return res
}
