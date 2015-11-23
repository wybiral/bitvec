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
)

func TestNotP10(t *testing.T) {
	b := NewBitvec()
	n := 100000
	p := 0.10
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
		t.Errorf("Incorrect count for Not")
	}
}

func TestAndP10(t *testing.T) {
	b1 := NewBitvec()
	b2 := NewBitvec()
	n := 100000
	p := 0.10
	count := 0
	for i := 0; i < n; i++ {
		x1 := rand.Float64() < p
		x2 := rand.Float64() < p
		if x1 {
			b1.Set(i, true)
		}
		if x2 {
			b2.Set(i, true)
		}
		if x1 && x2 {
			count++
		}
	}
	if count != b1.Iterate().And(b2.Iterate()).Count() {
		t.Errorf("Incorrect count for And")
	}
}

func TestOrP10(t *testing.T) {
	b1 := NewBitvec()
	b2 := NewBitvec()
	n := 100000
	p := 0.10
	count := 0
	for i := 0; i < n; i++ {
		x1 := rand.Float64() < p
		x2 := rand.Float64() < p
		if x1 {
			b1.Set(i, true)
		}
		if x2 {
			b2.Set(i, true)
		}
		if x1 || x2 {
			count++
		}
	}
	if count != b1.Iterate().Or(b2.Iterate()).Count() {
		t.Errorf("Incorrect count for Or")
	}
}

func TestXorP10(t *testing.T) {
	b1 := NewBitvec()
	b2 := NewBitvec()
	n := 100000
	p := 0.10
	count := 0
	for i := 0; i < n; i++ {
		x1 := rand.Float64() < p
		x2 := rand.Float64() < p
		if x1 {
			b1.Set(i, true)
		}
		if x2 {
			b2.Set(i, true)
		}
		if (x1 || x2) && !(x1 && x2) {
			count++
		}
	}
	if count != b1.Iterate().Xor(b2.Iterate()).Count() {
		t.Errorf("Incorrect count for Xor")
	}
}

func TestBitvecP01(t *testing.T) {
	b := NewBitvec()
	n := 100000
	p := 0.01
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
			return
		}
	}
}

func TestBitvecP50(t *testing.T) {
	b := NewBitvec()
	n := 100000
	p := 0.50
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
			return
		}
	}
}
