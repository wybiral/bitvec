package bitvec

type Iterator interface {
	Next() (val Word, ok bool)
}

type notItr struct {
	x Iterator
}

func (s *notItr) Next() (Word, bool) {
	val, ok := s.x.Next()
	return ^(val | FILL_BIT), ok
}

func Not(x Iterator) Iterator {
	return &notItr{x}
}

type binaryItr struct {
	x, y Iterator
}

type andItr binaryItr
type orItr binaryItr
type xorItr binaryItr

func (s *andItr) Next() (Word, bool) {
	x, xok := s.x.Next()
	y, yok := s.y.Next()
	return x & y, xok && yok
}

func And(x, y Iterator) Iterator {
	return &andItr{x, y}
}

func (s *orItr) Next() (Word, bool) {
	x, xok := s.x.Next()
	y, yok := s.y.Next()
	return x | y, xok && yok
}

func Or(x, y Iterator) Iterator {
	return &orItr{x, y}
}

func (s *xorItr) Next() (Word, bool) {
	x, xok := s.x.Next()
	y, yok := s.y.Next()
	return x ^ y, xok && yok
}

func Xor(x, y Iterator) Iterator {
	return &xorItr{x, y}
}

func Count(s Iterator) int {
	count := 0
	for {
		x, ok := s.Next()
		if !ok {
			break
		}
		for x > 0 {
			count++
			x &= (x - 1)
		}
	}
	return count
}

func Ids(s Iterator) chan int {
	ch := make(chan int)
	go func() {
		id := 0
		for {
			w, ok := s.Next()
			if !ok {
				break
			}
			for i := Word(0); i < Wordbits-1; i++ {
				if w&(1<<i) != 0 {
					ch <- id + int(i)
				}
			}
			id += Wordbits - 1
		}
		close(ch)
	}()
	return ch
}
