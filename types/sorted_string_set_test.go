package types

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_Less(t *testing.T) {
	s1 := String{"test"}
	s2 := String{"test"}
	assert.False(t, s1.Less(s2))
	assert.False(t, s2.Less(s1))
	assert.True(t, !s1.Less(s2) && !s2.Less(s1))
	//^ necessary for a string weak ordering
	s1 = String{"A"}
	s1 = String{"B"}
	assert.True(t, s1.Less(s2))
}

func TestSSSIteration(t *testing.T) {
	set := NewSortedStringSet()
	var expected []string
	for c := rune(65); c < 123; c++ {
		set.Add(string(c))
		expected = append(expected, string(c))
	}
	var got []string
	_ = set.ForEach(func(s string) error {
		got = append(got, s)
		return nil
	})
	sort.Strings(expected)
	assert.Equal(t, expected, got)
}

func TestSortedStringSet_Delete(t *testing.T) {
	set := NewSortedStringSet()
	set.Add("f")
	set.Add("d")
	set.Add("e")
	set.Add("c")
	set.Add("a")
	set.Add("b")
	set.Delete("a")
	assert.False(t, set.Contains("a"))
	for i := 98; i < 103; i++ {
		set.Delete(string(rune(i)))
	}
	assert.Equal(t, 0, set.Size())
}
