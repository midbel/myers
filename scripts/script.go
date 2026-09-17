package main

import (
	"fmt"
	"io"
	"os"

	"github.com/midbel/myers"
)

type RuneFormatter struct{}

func (f RuneFormatter) Equal(w io.Writer, value []rune) error {
	return f.formatString(w, value)
}

func (f RuneFormatter) Insert(w io.Writer, value []rune) error {
	return f.formatString(w, value)
}

func (f RuneFormatter) Delete(w io.Writer, value []rune) error {
	return f.formatString(w, value)
}

func (f RuneFormatter) formatString(w io.Writer, value []rune) error {
	_, err := io.WriteString(w, string(value))
	return err
}

func main() {
	fst := "The quick brown fox jumps over the lazy dog while the small birds sing softly beside the old oak tree and the wind moves gently through the green leaves carrying distant sounds from the quiet village near the river where children play and families gather during the warm summer evenings under the clear blue sky as stars begin to appear above the hills and the moon slowly rises over the peaceful valley where everything seems calm and unchanged except for the sudden noise coming from the market square where merchants are closing their shops and preparing their goods for tomorrow morning when the streets will once again become busy with people walking toward the station carrying bags books and newspapers before catching their trains and returning home to their families who are waiting patiently at the end of another ordinary day filled with small events and familiar conversations that make the whole town feel alive and connected despite the darkness spreading across the rooftops and gardens while the distant bells announce the late hour and everyone finally settles down for the night."
	snd := "The quick brown fox jumps over the lazy dog while the small birds sing softly beside the old oak tree and the wind moves gently through the green leaves carrying distant sounds from the quiet village near the river where children play and families gather during the warm summer evenings under the clear blue sky as stars begin to appear above the hills and the moon slowly rises over the peaceful valley where everything seems calm and unchanged except for the sudden noise coming from the market square where merchants are closing their shops and preparing their goods for tomorrow morning when the streets will once again become crowded with people walking toward the station carrying bags letters and newspapers before catching their trains and returning home to their families who are waiting patiently at the end of another ordinary day filled with small events and familiar conversations that make the whole town feel alive and connected despite the darkness spreading across the rooftops and gardens while the distant bells announce the late hour and everyone finally settles down for the night."

	equal := func(a, b rune) bool {
		return a == b
	}

	var formatter RuneFormatter

	err := myers.DiffFunc(os.Stdout, formatter, []rune(fst), []rune(snd), equal)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
