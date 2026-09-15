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

func main() {
	flag.Parse()

	res1, err := Lines(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	res2, err := Lines(flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ops := myers.Script(res1, res2)
	for i := range ops {
		fmt.Println(ops[i])
	}
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
