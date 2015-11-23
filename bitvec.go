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

type Word uint32

const Wordbits = 32

const (
	FILL_BIT      = Word(1 << (Wordbits - 1))       // Mask for fill flag
	ONES_BIT      = Word(1 << (Wordbits - 2))       // Mask for ones flag
	FILL_MAX      = Word((2 << (Wordbits - 3)) - 1) // Maximum fill count
	COUNT_BITS    = ^(FILL_BIT | ONES_BIT)          // Mask for fill count bits
	ONES_LITERAL  = ^FILL_BIT                       // Filled ones literal
	ZEROS_LITERAL = Word(0)                         // Filled zeros literal
)

func isZerosFill(x Word) bool {
	return x & ^COUNT_BITS == FILL_BIT
}

func isOnesFill(x Word) bool {
	return x & ^COUNT_BITS == ^COUNT_BITS
}

func hasSpace(x Word) bool {
	return x&COUNT_BITS < FILL_MAX
}

type Bitvec struct {
	size   int    // Number of bits used (zero and one)
	active Word   // Currently active Word
	offset Word   // Which bit we're at in the active Word
	words  []Word // Allocated words
}

func NewBitvec() *Bitvec {
	return &Bitvec{size: 0, active: Word(0), offset: Word(0), words: make([]Word, 0, 64)}
}

func (b *Bitvec) append(x bool) {
	if x {
		b.active |= 1 << b.offset
	}
	b.offset++
	b.size++
	if b.offset == Wordbits-1 {
		b.flushWord()
	}
}

func (b *Bitvec) flushWord() {
	top := len(b.words) - 1
	if b.active == ZEROS_LITERAL {
		// All zero literal
		if top > -1 && isZerosFill(b.words[top]) && hasSpace(b.words[top]) {
			b.words[top]++
		} else {
			b.words = append(b.words, FILL_BIT)
		}
	} else if b.active == ONES_LITERAL {
		// All one literal
		if top > -1 && isOnesFill(b.words[top]) && hasSpace(b.words[top]) {
			b.words[top]++
		} else {
			b.words = append(b.words, FILL_BIT|ONES_BIT)
		}
	} else {
		b.words = append(b.words, b.active)
	}
	b.offset = Word(0)
	b.active = Word(0)
}

func (b *Bitvec) Set(id int, x bool) bool {
	for id > b.size {
		// There's a better way to do this, but for now...
		// Grow the vector one-by-one.
		b.append(false)
	}
	if id == b.size {
		b.append(x)
		return x
	}
	return b.update(id, x)
}

func (b *Bitvec) update(id int, x bool) bool {
	index := id / (Wordbits - 1)
	offset := Word(id % (Wordbits - 1))
	i := 0
	j := 0
	n := len(b.words)
	for ; i < n; i++ {
		nj := 1
		if b.words[i]&FILL_BIT != 0 {
			nj += int(b.words[i] & COUNT_BITS)
		}
		if j+nj > index {
			break
		}
		j += nj
	}
	if i == n {
		// Modify active Word
		old := b.active
		if x {
			b.active |= 1 << offset
		} else {
			b.active &= ^(1 << offset)
		}
		return old != b.active
	} else if b.words[i]&FILL_BIT != 0 {
		// Modify fill Word
		if (x && b.words[i]&ONES_BIT == 0) || !(x || b.words[i]&ONES_BIT == 0) {
			// Break this fill
			b.updateFill(i, Word(index-j), offset, x)
			return true
		}
	} else {
		// Modify literal Word
		return b.updateLiteral(i, offset, x)
	}
	return false
}

func (b *Bitvec) updateFill(i int, target Word, offset Word, x bool) {
	head := b.words[i] & (FILL_BIT | ONES_BIT)
	size := b.words[i] & COUNT_BITS
	if target > 0 {
		// There's a fill before the literal we're adding
		b.words[i] = head | (target - 1)
		b.words = append(b.words, 0)
		i++
		copy(b.words[i+1:], b.words[i:])
	}
	// Add the literal
	if x {
		b.words[i] = (1 << offset)
	} else {
		b.words[i] = (^FILL_BIT) ^ (1 << offset)
	}
	if size > target {
		// There's a fill after the literal
		b.words = append(b.words, 0)
		i++
		copy(b.words[i+1:], b.words[i:])
		b.words[i] = head | ((size - target) - 1)
	}
}

func (b *Bitvec) updateLiteral(i int, offset Word, x bool) bool {
	old := b.words[i]
	if x {
		b.words[i] |= 1 << offset
		if b.words[i] == ONES_LITERAL {
			if i > 0 && isOnesFill(b.words[i-1]) && hasSpace(b.words[i-1]) {
				b.words[i-1]++
				n := len(b.words) - 1
				copy(b.words[i:], b.words[i+1:])
				b.words[n] = Word(0)
				b.words = b.words[:n]
			} else {
				b.words[i] = FILL_BIT | ONES_BIT
			}
		}
	} else {
		b.words[i] &= ^(1 << offset)
		if b.words[i] == ZEROS_LITERAL {
			if i > 0 && isZerosFill(b.words[i-1]) && hasSpace(b.words[i-1]) {
				b.words[i-1]++
				n := len(b.words) - 1
				copy(b.words[i:], b.words[i+1:])
				b.words[n] = Word(0)
				b.words = b.words[:n]
			} else {
				b.words[i] = FILL_BIT
			}
		}
	}
	return b.words[i] != old
}

func (b *Bitvec) Get(id int) bool {
	index := id / (Wordbits - 1)
	offset := Word(id % (Wordbits - 1))
	i := 0
	j := 0
	n := len(b.words)
	for ; i < n; i++ {
		nj := 1
		if b.words[i]&FILL_BIT != 0 {
			nj += int(b.words[i] & COUNT_BITS)
		}
		if j+nj > index {
			break
		}
		j += nj
	}
	if i == n {
		return b.active&(1<<offset) != 0
	} else if b.words[i]&FILL_BIT != 0 {
		return b.words[i]&ONES_BIT != 0
	}
	return b.words[i]&(1<<offset) != 0
}

type iterator struct {
	b    *Bitvec
	i    int
	n    Word
	fill Word
}

func (b *Bitvec) Iterate() *Iterator {
	length := (b.size / (Wordbits - 1)) + 1
	offset := b.offset
	itr := &Iterator{nil, length, offset}
	i := 0
	n := Word(0)
	fill := Word(0)
	itr.next = func() Word {
		if n > 0 {
			n--
			return fill
		}
		if i >= len(b.words) {
			if i == len(b.words) {
				i++
				return b.active
			}
			return 0
		}
		w := b.words[i]
		if w&FILL_BIT == 0 {
			i++
			return w
		} else {
			i++
			n = (w & COUNT_BITS)
			if w&ONES_BIT == 0 {
				fill = 0
			} else {
				fill = ^FILL_BIT
			}
			return fill
		}
	}
	return itr
}
