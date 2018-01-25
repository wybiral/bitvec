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

type Iterator struct {
	next   func() Word // Function returning next word
	length int         // Length in number of words
	offset Word        // Offset bit in active word
}

func ZeroIterator() *Iterator {
	next := func() Word {
		return Word(0)
	}
	return &Iterator{next, 0, 0}
}

func (x *Iterator) Not() *Iterator {
	index := 0
	next := func() Word {
		val := ^x.next()
		index++
		if index == x.length {
			val &= (^Word(0)) >> (bitLength - x.offset)
		}
		return val & ^fillFlag
	}
	return &Iterator{next, x.length, x.offset}
}

func (x *Iterator) And(y *Iterator) *Iterator {
	index := 0
	length := x.length
	offset := x.offset
	if y.length < length {
		length = y.length
		offset = y.offset
	} else if y.offset < offset {
		offset = y.offset
	}
	itr := &Iterator{nil, length, offset}
	itr.next = func() Word {
		xval := x.next()
		yval := y.next()
		index++
		if index == length {
			// All zero after this point
			itr.next = func() Word {
				return 0
			}
		}
		return xval & yval & ^fillFlag
	}
	return itr
}

func (x *Iterator) Or(y *Iterator) *Iterator {
	index := 0
	length := x.length
	offset := x.offset
	if y.length > length {
		length = y.length
		offset = y.offset
	} else if y.offset > offset {
		offset = y.offset
	}
	itr := &Iterator{nil, length, offset}
	itr.next = func() Word {
		xval := x.next()
		yval := y.next()
		index++
		return (xval | yval) & ^fillFlag
	}
	return itr
}

func (x *Iterator) Xor(y *Iterator) *Iterator {
	index := 0
	length := x.length
	offset := x.offset
	if y.length > length {
		length = y.length
		offset = y.offset
	} else if y.offset > offset {
		offset = y.offset
	}
	itr := &Iterator{nil, length, offset}
	itr.next = func() Word {
		xval := x.next()
		yval := y.next()
		index++
		return (xval ^ yval) & ^fillFlag
	}
	return itr
}

// Sparse bit counter, is there a better option?
func (itr *Iterator) Count() int {
	count := 0
	for i := 0; i < itr.length; i++ {
		count += bits.OnesCount(uint(itr.next()))
	}
	return count
}

// Return channel of Ids for 1 bits
func (itr *Iterator) Ids() chan int {
	ch := make(chan int)
	go func() {
		id := 0
		for i := 0; i < itr.length; i++ {
			w := itr.next()
			for i := Word(0); i < bitLength-1; i++ {
				if w&(1<<i) != 0 {
					ch <- id + int(i)
				}
			}
			id += bitLength - 1
		}
		close(ch)
	}()
	return ch
}
