package types

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

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

func BenchmarkSortedStringSet_Add(b *testing.B) {
	choice := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRTSTUVWXYZ1234567890"
	set := NewSortedStringSet()
	start := time.Now()
	for i := 0; i < 100000; i++ {
		s := strings.Builder{}
		for j := 0; j < 70; j++ {
			s.WriteByte(choice[rand.Int() % len(choice)])
		}
		set.Add(s.String())
	}
	fmt.Println(time.Since(start).Milliseconds(), "ms to add")
	start = time.Now()
	var iterTimes []int64
	iterStart := time.Now()
	_ = set.ForEach(func(s string) error {
		iterTimes = append(iterTimes, time.Since(iterStart).Nanoseconds())
		iterStart = time.Now()
		return nil
	})
	fmt.Println(time.Since(start).Milliseconds(), "ms to iterate")
	max, min, sum := int64(0), int64(99999999), int64(0)
	maxi := 0
	for i, val := range iterTimes {
		if val > max {
			max = val
			maxi = i
		}
		if val < min {
			min = val
		}
		sum += val
	}
	avg := 		sum / int64(len(iterTimes))
		fmt.Printf(
		"Max: %vns @ %v%% of the way through, min: %vns, sum: %vns.\n",
		max,
		(float64(maxi) / float64(len(iterTimes))) * 100,
		min,
		avg,
	)
	fmt.Printf("max is %vx larger than sum, %vx larger than min\n\n",
		float64(max)/float64(avg),
		float64(max)/float64(min))
}
