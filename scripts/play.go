package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/midbel/myers"
)

type str string

func (c str) Equal(other str) bool {
	return c == other
}

type ArraySeq[T myers.Equaler[T]] struct {
	array  []T
	offset int
}

func NewSequenceFromFile(file string) (myers.Sequence[str], error) {
	lines, err := Lines(file)
	if err != nil {
		return nil, err
	}
	seq := &ArraySeq[str]{
		array:  lines,
		offset: 0,
	}
	return seq, nil
}

func (s *ArraySeq[T]) Peek() (zero T, err error) {
	if s.offset >= len(s.array) {
		err = myers.End
		return zero, err
	}
	return s.array[s.offset], nil
}

func (s *ArraySeq[T]) Next() (T, error) {
	z, err := s.Peek()
	if err == nil {
		s.offset++
	}
	return z, err
}

func (s *ArraySeq[T]) Fork() myers.Sequence[T] {
	x := &ArraySeq[T]{
		offset: s.offset,
		array:  s.array,
	}
	return x
}

func main() {
	flag.Parse()

	res1, err := NewSequenceFromFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	res2, err := NewSequenceFromFile(flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	myers.Explore(res1, res2, func(entry myers.Entry[str]) error {
		fmt.Println(entry.Edit, entry.Op, entry.First, entry.Second, entry.Success, entry.Failure)
		return nil
	})
}

func Lines(file string) ([]str, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var (
		scan  = bufio.NewScanner(r)
		lines []str
	)
	for scan.Scan() {
		lines = append(lines, str(scan.Text()))
	}
	return lines, scan.Err()
}
