package types

type StringSet struct {
    internal map[string]struct{}
}

func NewThreadSafeSet() *StringSet {
    return  &StringSet{
        internal: make(map[string]struct{}),
    }
}

func (tss *StringSet) Contains(str string) bool {
    _, ok := tss.internal[str]
    return ok
}

func (tss *StringSet) Add(str string) {
    tss.internal[str] = struct{}{}
}

func (tss *StringSet) Delete(str string) {
    delete(tss.internal, str)
}

func (tss *StringSet) Size() int {
    return len(tss.internal)
}
