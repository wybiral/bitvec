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

// This package implements compressed bitvectors using the WAH method.

package bitvec

type Word uint64

// There are two types of words denoted by the highest bit known as the "fill
// flag" or "signal bit".
//
// Literal words have a fill flag of 0 and the remaining bits are to be taken
// literally as the bit pattern for that word.
//
// Fill words have a fill flag of 1 and the next highest bit is the "ones flag"
// or "pattern bit". If this bit is set to 0 then this is a 0-fill word and if
// it's set to 1 then this is a 1-fill word. The remaining bits are used as an
// integer count for the number of repetitions for this fill.

const (
	bitLength    = 64                               // Bits per word
	fillFlag     = Word(1 << (bitLength - 1))       // Mask for fill flag
	onesFlag     = Word(1 << (bitLength - 2))       // Mask for ones flag
	fillMax      = Word((2 << (bitLength - 3)) - 1) // Maximum fill count
	countMask    = ^(fillFlag | onesFlag)           // Mask for fill count bits
	onesLiteral  = ^fillFlag                        // Filled ones literal
	zerosLiteral = Word(0)                          // Filled zeros literal
)

// Is this word a fill of zeros?
func isZerosFill(x Word) bool {
	return x & ^countMask == fillFlag
}

// Is this word a fill of ones?
func isOnesFill(x Word) bool {
	return x & ^countMask == ^countMask
}

// Can this fill count be incremented without overflowing?
func hasSpace(x Word) bool {
	return x&countMask < fillMax
}

type Bitvec struct {
	size   int    // Number of bits used (zero and one)
	active Word   // Currently active Word
	offset int    // Which bit we're at in the active Word
	words  []Word // Allocated words
}

// Return a new *BitVec of size 0
func New() *Bitvec {
	return &Bitvec{
		size:   0,
		active: Word(0),
		offset: 0,
		words:  make([]Word, 0),
	}
}

// Append boolean bit to vector
func (b *Bitvec) append(x bool) {
	if x {
		b.active |= 1 << uint(b.offset)
	}
	b.offset++
	b.size++
	if b.offset == bitLength-1 {
		b.flushWord()
	}
}

// Adds current active word to rest of vector and creates new active word
func (b *Bitvec) flushWord() {
	top := len(b.words) - 1
	if b.active == zerosLiteral {
		// All zero literal
		if top > -1 && isZerosFill(b.words[top]) && hasSpace(b.words[top]) {
			b.words[top]++
		} else {
			b.words = append(b.words, fillFlag)
		}
	} else if b.active == onesLiteral {
		// All one literal
		if top > -1 && isOnesFill(b.words[top]) && hasSpace(b.words[top]) {
			b.words[top]++
		} else {
			b.words = append(b.words, fillFlag|onesFlag)
		}
	} else {
		b.words = append(b.words, b.active)
	}
	b.active = Word(0)
	b.offset = 0
}

// Set bit at id, expanding as needed
func (b *Bitvec) Set(id int, x bool) {
	if id > b.size {
		offset := b.offset + id - b.size
		words := offset / (bitLength - 1)
		for i := 0; i < words; i++ {
			b.flushWord()
		}
		b.offset = offset % (bitLength - 1)
		b.size = id
	}
	if id == b.size {
		// id is just after the end of Bitvec so append
		b.append(x)
		return
	}
	// Not an append, handle update
	b.update(id, x)
}

func (b *Bitvec) update(id int, x bool) {
	index, offset, i, j := b.findWord(id)
	if i == len(b.words) {
		// Modify active Word
		if x {
			b.active |= 1 << uint(offset)
		} else {
			b.active &= ^(1 << uint(offset))
		}
	} else if b.words[i]&fillFlag != 0 {
		// Modify fill Word
		if (x && b.words[i]&onesFlag == 0) || !(x || b.words[i]&onesFlag == 0) {
			// x doesn't match fill type, break this fill
			b.updateFill(i, Word(index-j), offset, x)
		}
	} else {
		// Modify literal Word
		b.updateLiteral(i, offset, x)
	}
}

func (b *Bitvec) updateFill(i int, target Word, offset int, x bool) {
	head := b.words[i] & (fillFlag | onesFlag)
	size := b.words[i] & countMask
	if target > 0 {
		// There's a fill before the literal we're adding
		b.words[i] = head | (target - 1)
		b.words = append(b.words, 0)
		i++
		copy(b.words[i+1:], b.words[i:])
	}
	// Add the literal
	if x {
		b.words[i] = (1 << uint(offset))
	} else {
		b.words[i] = (^fillFlag) ^ (1 << uint(offset))
	}
	if size > target {
		// There's a fill after the literal
		b.words = append(b.words, 0)
		i++
		copy(b.words[i+1:], b.words[i:])
		b.words[i] = head | ((size - target) - 1)
	}
}

func (b *Bitvec) updateLiteral(i int, offset int, x bool) {
	if x {
		b.words[i] |= 1 << uint(offset)
		if b.words[i] == onesLiteral {
			// Our update made this literal a fill...
			if i > 0 && isOnesFill(b.words[i-1]) && hasSpace(b.words[i-1]) {
				// Previous word is matching fill with space to increment
				b.words[i-1]++
				n := len(b.words) - 1
				copy(b.words[i:], b.words[i+1:])
				b.words[n] = Word(0)
				b.words = b.words[:n]
			} else {
				b.words[i] = fillFlag | onesFlag
			}
		}
	} else {
		b.words[i] &= ^(1 << uint(offset))
		if b.words[i] == zerosLiteral {
			// Our update made this literal a fill...
			if i > 0 && isZerosFill(b.words[i-1]) && hasSpace(b.words[i-1]) {
				// Previous word is matching fill with space to increment
				b.words[i-1]++
				n := len(b.words) - 1
				copy(b.words[i:], b.words[i+1:])
				b.words[n] = Word(0)
				b.words = b.words[:n]
			} else {
				b.words[i] = fillFlag
			}
		}
	}
}

// Find which word a bit is located in.
// index is the word index assuming all words are literal (ignores fills)
// offset is the bit offset
// i is the actual index into word array (can extend beyond array length)
// j is the index of the previous word in the array (ignores the last fill)
func (b *Bitvec) findWord(id int) (index, offset, i, j int) {
	index = id / (bitLength - 1)
	offset = id % (bitLength - 1)
	n := len(b.words)
	for ; i < n; i++ {
		nextj := j + 1
		if b.words[i]&fillFlag != 0 {
			nextj += int(b.words[i] & countMask)
		}
		if nextj > index {
			return
		}
		j = nextj
	}
	// Correct for OOB id
	if j + 1 < index {
		i++
	}
	return
}

// Get state of bit at id
func (b *Bitvec) Get(id int) bool {
	_, offset, i, _ := b.findWord(id)
	if i == len(b.words) {
		return b.active&(1<<uint(offset)) != 0
	}
	if i > len(b.words) {
		return false
	}
	if b.words[i]&fillFlag != 0 {
		return b.words[i]&onesFlag != 0
	}
	return b.words[i]&(1<<uint(offset)) != 0
}

func (b *Bitvec) Iterate() Iterator {
	return &bitvecIterator{b: b}
}
