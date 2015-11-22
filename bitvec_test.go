package bitvec

import (
	"testing"
	"math/rand"
)

func TestBitvecP01(t *testing.T) {
	b := NewBitvec()
	n := 1000000
	p := 0.01
	data := make([]int, 0)
	for i := 0; i < n; i++ {
		if rand.Float64() < p {
			data = append(data, i)
			b.Set(i, true)
		}
	}
	ids := Ids(b.Iterate())
	for i, x := range data {
		y := <- ids
		if x != y {
			t.Errorf("Failed on the %dth value", i)
			return
		}
	}
}

func TestBitvecP50(t *testing.T) {
	b := NewBitvec()
	n := 1000000
	p := 0.50
	data := make([]int, 0)
	for i := 0; i < n; i++ {
		if rand.Float64() < p {
			data = append(data, i)
			b.Set(i, true)
		}
	}
	ids := Ids(b.Iterate())
	for i, x := range data {
		y := <- ids
		if x != y {
			t.Errorf("Failed on the %dth value", i)
			return
		}
	}
}
