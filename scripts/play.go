package main

import (
	"fmt"

	"github.com/midbel/myers"
)

type char rune

func (c char) Equal(other char) bool {
	return c == other
}

func chars(str string) []char {
	list := make([]char, 0, len(str))
	for _, r := range str {
		list = append(list, char(r))
	}
	return list
}

func main() {
	var (
		fst = chars("abcabba")
		snd = chars("cbabac")
	)
	myers.Explore(fst, snd, func(edit int, op myers.Op, a, b char, success, early bool) {
		fmt.Println(edit, string(a), string(b))
	})
	// fmt.Println(ops)
}
