package string_set

import (
	"math/rand"
	"runtime"
	"strconv"
	"testing"
)

func testSimpleAdds(t *testing.T, set StringSet) {
	if set.Count() != 0 {
		t.Errorf("Expected count to be 0")
	}
	if !set.Add("a") {
		t.Errorf("Expected insert of a to succeed")
	}
	if set.Add("a") {
		t.Errorf("Expected insert of a to fail, already inserted")
	}
	if set.Count() != 1 {
		t.Errorf("Count should be 1 after inserting a twice")
	}

	if !set.Add("b") {
		t.Errorf("Expected insert of b to succeed")
	}
	if set.Count() != 2 {
		t.Errorf("Expected count to update to 2")
	}
}

func testConcurrentAdds(t *testing.T, set StringSet) {
	ch := make(chan struct{})
	finalCh := make(chan struct{})

	numGoros := runtime.NumCPU()
	expectedNumItems := ((numGoros-1)/5 + 1) * 1000
	for i := 0; i < numGoros; i++ {
		go func(i int) {
			for x := 0; x < 100; x++ {
				for y := 0; y < 1000; y++ {
					set.Add(strconv.Itoa(i/5 + y*1000))
				}
			}
			ch <- struct{}{}
		}(i)
		// spinnup readers to continue reading count()
		go func(i int) {
			for {
				select {
				case <-finalCh:
					return
				default:
					if c := set.Count(); c > expectedNumItems {
						t.Errorf("Expected %d items, got %d", expectedNumItems, c)
					}
				}
			}
		}(i)
	}
	for i := 0; i < numGoros; i++ {
		<-ch
	}
	// stop readers
	close(finalCh)

	if c := set.Count(); c != expectedNumItems {
		t.Errorf("Expected %d items, got %d", expectedNumItems, c)
	}
}

func TestLockedSimpleAdds(t *testing.T) {
	set := MakeLockedStringSet()
	testSimpleAdds(t, &set)
}

func TestLockedConcurrentAdds(t *testing.T) {
	set := MakeLockedStringSet()
	testConcurrentAdds(t, &set)
}

func TestStripedSimpleAdds(t *testing.T) {
	for stripes := 1; stripes <= 20; stripes++ {
		set := MakeStripedStringSet(stripes)
		testSimpleAdds(t, &set)
	}
}

func TestStripedConcurrentAdds(t *testing.T) {
	for stripes := 1; stripes <= 20; stripes++ {
		set := MakeStripedStringSet(stripes)
		testConcurrentAdds(t, &set)
	}
}

func benchmarkAdds(b *testing.B, set StringSet) {
	b.RunParallel(func(pb *testing.PB) {
		i := rand.Intn(10)
		for j := 0; pb.Next(); j++ {
			set.Add(strconv.Itoa(i + j%100))
		}
	})
}

func BenchmarkLockedStringSet(b *testing.B) {
	set := MakeLockedStringSet()
	benchmarkAdds(b, &set)
}

func BenchmarkStripedStringSet1(b *testing.B) {
	set := MakeStripedStringSet(1)
	benchmarkAdds(b, &set)
}

func BenchmarkStripedStringSet2(b *testing.B) {
	set := MakeStripedStringSet(2)
	benchmarkAdds(b, &set)
}

func BenchmarkStripedStringSetNumCores(b *testing.B) {
	set := MakeStripedStringSet(runtime.NumCPU())
	benchmarkAdds(b, &set)
}

func BenchmarkStripedStringSetNumCoresTimesTwo(b *testing.B) {
	set := MakeStripedStringSet(runtime.NumCPU() * 2)
	benchmarkAdds(b, &set)
}

func BenchmarkStripedStringSetNumCoresTimesEight(b *testing.B) {
	set := MakeStripedStringSet(runtime.NumCPU() * 8)
	benchmarkAdds(b, &set)
}
