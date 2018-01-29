// Copyright 2015 Davy Wybiral <davy.wybiral@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bitvec

import (
	"math/bits"
)

// Iteration is done using a Next() method that returns a literal word and the
// number of bits it represents.
// For partial literals that means the number will be less than bitLength - 1
// since the fillFlag bit isn't counted.
// Iteration is complete when the number of bits returned is 0.
type Iterator interface {
	Next() (Word, int)
}

// Empty iterator
type emptyIterator struct {}

func (itr emptyIterator) Next() (Word, int) {
	return Word(0), 0
}

func EmptyIterator() Iterator {
	return emptyIterator{}
}

// Zero iterator
type zeroIterator struct {
	n int
}

func (itr zeroIterator) Next() (Word, int) {
	n := bitLength - 1
	if itr.n >= n {
		itr.n -= n
	} else {
		n = itr.n
		itr.n = 0
	}
	return Word(0), n
}

func ZeroIterator(n int) Iterator {
	return zeroIterator{n: n}
}

// Iterator implementation for bitvectors
type bitvecIterator struct {
	b     *Bitvec // bitvector being iterated
	index int     // current encoded word
	count int     // count used for fill runs
	fill  Word    // fill word used for fill runs
}

func (itr *bitvecIterator) Next() (Word, int) {
	// Iterating fill count
	if itr.count > 0 {
		itr.count--
		return itr.fill, bitLength - 1
	}
	if itr.index < len(itr.b.words) {
		w := itr.b.words[itr.index]
		itr.index++
		// Literal word
		if w&fillFlag == 0 {
			return w, bitLength - 1
		}
		// Fill word
		itr.count = int(w & countMask)
		if w&onesFlag == 0 {
			itr.fill = 0
		} else {
			itr.fill = ^fillFlag
		}
		return itr.fill, bitLength - 1
	}
	// Active (partial) literal word
	if itr.index == len(itr.b.words) {
		itr.index++
		return itr.b.active, itr.b.offset
	}
	// End of stream
	return Word(0), 0
}

// Bitwise NOT iterator.
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

// Bitwise AND iterator.
type andIterator struct {
	x Iterator
	y Iterator
}

func (itr *andIterator) Next() (Word, int) {
	wx, nx := itr.x.Next()
	wy, ny := itr.y.Next()
	return wx & wy, min(nx, ny)
}

func And(x, y Iterator) Iterator {
	return &andIterator{x, y}
}

// Bitwise OR iterator.
type orIterator struct {
	x Iterator
	y Iterator
}

func (itr *orIterator) Next() (Word, int) {
	wx, nx := itr.x.Next()
	wy, ny := itr.y.Next()
	return wx | wy, max(nx, ny)
}

func Or(x, y Iterator) Iterator {
	return &orIterator{x, y}
}

// Bitwise XOR iterator.
type xorIterator struct {
	x Iterator
	y Iterator
}

func (itr *xorIterator) Next() (Word, int) {
	wx, nx := itr.x.Next()
	wy, ny := itr.y.Next()
	return wx ^ wy, max(nx, ny)
}

func Xor(x, y Iterator) Iterator {
	return &xorIterator{x, y}
}

// Count all bits set to 1 in iterator.
func Count(itr Iterator) int {
	const mask = ^Word(0)
	count := 0
	for {
		w, n := itr.Next()
		if n == 0 {
			break
		}
		if n < bitLength-1 {
			w &= mask >> uint(bitLength-n)
		}
		count += bits.OnesCount(uint(w))
	}
	return count
}

// Return channel of integer indices of bits set to 1 in iterator.
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
				if w&(1<<uint(i)) != 0 {
					ch <- id + i
				}
			}
			id += bitLength - 1
		}
		close(ch)
	}()
	return ch
}

// Return minimum of x and y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Return maximum of x and y.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
