/*
Copyright 2015 Davy Wybiral <davy.wybiral@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bitvec

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func randomTest(fn func(n int, p float64) bool) {
	sizes := []int{100, 10000, 100000, 1000000}
	ratios := []float64{0.1, 0.25, 0.5, 0.75, 0.9}
	for _, n := range sizes {
		for _, p := range ratios {
			if !fn(n, p) {
				return
			}
		}
	}
}

func TestNot(t *testing.T) {
	randomTest(func(n int, p float64) bool {
		b := New()
		count := 0
		for i := 0; i < n; i++ {
			if rand.Float64() < p {
				b.Set(i, true)
			} else {
				b.Set(i, false)
				count++
			}
		}
		if count != b.Iterate().Not().Count() {
			t.Errorf("Incorrect count for Not, n=%d, p=%f", n, p)
			return false
		}
		return true
	})
}

func TestAnd(t *testing.T) {
	randomTest(func(n int, p float64) bool {
		b1 := New()
		b2 := New()
		count := 0
		for i := 0; i < n; i++ {
			x1 := rand.Float64() < p
			x2 := rand.Float64() < p
			b1.Set(i, x1)
			b2.Set(i, x2)
			if x1 && x2 {
				count++
			}
		}
		if count != b1.Iterate().And(b2.Iterate()).Count() {
			t.Errorf("Incorrect count for And")
			return false
		}
		return true
	})
}

func TestOr(t *testing.T) {
	randomTest(func(n int, p float64) bool {
		b1 := New()
		b2 := New()
		count := 0
		for i := 0; i < n; i++ {
			x1 := rand.Float64() < p
			x2 := rand.Float64() < p
			b1.Set(i, x1)
			b2.Set(i, x2)
			if x1 || x2 {
				count++
			}
		}
		if count != b1.Iterate().Or(b2.Iterate()).Count() {
			t.Errorf("Incorrect count for Or")
			return false
		}
		return true
	})
}

func TestXor(t *testing.T) {
	randomTest(func(n int, p float64) bool {
		b1 := New()
		b2 := New()
		count := 0
		for i := 0; i < n; i++ {
			x1 := rand.Float64() < p
			x2 := rand.Float64() < p
			b1.Set(i, x1)
			b2.Set(i, x2)
			if x1 != x2 {
				count++
			}
		}
		if count != b1.Iterate().Xor(b2.Iterate()).Count() {
			t.Errorf("Incorrect count for Xor")
			return false
		}
		return true
	})
}

func TestIds(t *testing.T) {
	randomTest(func(n int, p float64) bool {
		b := New()
		data := make([]int, 0)
		for i := 0; i < n; i++ {
			if rand.Float64() < p {
				data = append(data, i)
				b.Set(i, true)
			}
		}
		ids := b.Iterate().Ids()
		for i, x := range data {
			y := <-ids
			if x != y {
				t.Errorf("Failed on the %dth value", i)
				return false
			}
		}
		return true
	})
}
