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
	"math/bits"
)

type Iterator interface {
	Next() (Word, int)
}

type notIterator struct {
	x Iterator
}

func (itr *notIterator) Next() (Word, int) {
	w, n := itr.x.Next()
	return ^fillFlag ^ w, n
}

func Not(x Iterator) Iterator {
	return &notIterator{x}
}


type andIterator struct {
	x Iterator
	y Iterator
}

func (itr *andIterator) Next() (Word, int) {
	var n int
	wx, nx := itr.x.Next()
	wy, ny := itr.y.Next()
	if nx < ny {
		n = nx
	} else {
		n = ny
	}
	return wx & wy, n
}

func And(x, y Iterator) Iterator {
	return &andIterator{x, y}
}


type orIterator struct {
	x Iterator
	y Iterator
}

func (itr *orIterator) Next() (Word, int) {
	var n int
	wx, nx := itr.x.Next()
	wy, ny := itr.y.Next()
	if nx < ny {
		n = nx
	} else {
		n = ny
	}
	return wx | wy, n
}

func Or(x, y Iterator) Iterator {
	return &orIterator{x, y}
}


type xorIterator struct {
	x Iterator
	y Iterator
}

func (itr *xorIterator) Next() (Word, int) {
	var n int
	wx, nx := itr.x.Next()
	wy, ny := itr.y.Next()
	if nx < ny {
		n = nx
	} else {
		n = ny
	}
	return wx ^ wy, n
}

func Xor(x, y Iterator) Iterator {
	return &xorIterator{x, y}
}


func Count(itr Iterator) int {
	count := 0
	for {
		w, n := itr.Next()
		if n == 0 {
			break
		}
		if n < bitLength - 1 {
			w &= (1 << uint(n)) - 1
		}
		count += bits.OnesCount(uint(w))
	}
	return count
}

func Indices(itr Iterator) chan int {
	ch := make(chan int)
	go func() {
		id := 0
		for {
			w, n := itr.Next()
			if n == 0 {
				break
			}
			for i := 0; i < n; i++ {
				if w & (1 << uint(i)) != 0 {
					ch <- id + int(i)
				}
			}
			id += bitLength - 1
		}
		close(ch)
	}()
	return ch
}
