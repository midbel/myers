package myers

import (
	"slices"
	"testing"
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

func TestScript(t *testing.T) {
	data := []struct {
		Fst  string
		Snd  string
		Want []Op
	}{
		{Fst: "abc", Snd: "abc"},
		{Fst: "", Snd: ""},
		{Fst: "abc", Snd: "", Want: []Op{DeleteOp, DeleteOp, DeleteOp}},
		{Fst: "", Snd: "abc", Want: []Op{InsertOp, InsertOp, InsertOp}},
		{Fst: "ab", Snd: "abc", Want: []Op{EqualOp, EqualOp, InsertOp}},
		{Fst: "abc", Snd: "ab", Want: []Op{EqualOp, EqualOp, DeleteOp}},
		{Fst: "bc", Snd: "abc", Want: []Op{InsertOp, EqualOp, EqualOp}},
		{Fst: "abc", Snd: "abd", Want: []Op{EqualOp, EqualOp, InsertOp, DeleteOp}},
		{Fst: "ab", Snd: "ba", Want: []Op{InsertOp, EqualOp, DeleteOp}},
		{
			Fst:  "abcabba",
			Snd:  "cbabac",
			Want: []Op{InsertOp, DeleteOp, EqualOp, DeleteOp, EqualOp, EqualOp, DeleteOp, EqualOp, InsertOp},
		},
	}
	for _, d := range data {
		got := Script(chars(d.Fst), chars(d.Snd))
		if !slices.Equal(got, d.Want) {
			t.Errorf("%q -> %q: unexpected script: want %v, got %v", d.Fst, d.Snd, d.Want, got)
		}
	}
}

func TestSame(t *testing.T) {
	data := []struct {
		Fst  string
		Snd  string
		Want bool
	}{
		{Fst: "abc", Snd: "abc", Want: true},
		{Fst: "", Snd: "", Want: true},
		{Fst: "abc", Snd: "abd"},
		{Fst: "abc", Snd: "ab"},
		{Fst: "", Snd: "abc"},
	}
	for _, d := range data {
		got := Same(chars(d.Fst), chars(d.Snd))
		if got != d.Want {
			t.Errorf("%q vs %q: want %t, got %t", d.Fst, d.Snd, d.Want, got)
		}
	}
}
