package main

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

var prefix = "Some prefix | "

func TestPrefixedLineReader(t *testing.T) {
	testStrings := []struct {
		in  string
		out string
	}{
		{
			in:  "",
			out: "",
		},
		{
			in:  "x",
			out: prefix + "x",
		},
		{
			in:  "\n",
			out: prefix + "\n",
		},
		{
			in:  "Hello World!\n",
			out: prefix + "Hello World!\n",
		},
		{
			in:  "Multiline\ntext",
			out: prefix + "Multiline\n" + prefix + "text",
		},
		{
			in:  "Multiline\ntext\nis\nawesome\n",
			out: prefix + "Multiline\n" + prefix + "text\n" + prefix + "is\n" + prefix + "awesome\n",
		},
	}

	for _, tt := range testStrings {
		r := NewPrefixedLineReader(strings.NewReader(tt.in), prefix)

		buf, err := ioutil.ReadAll(r)
		if err != nil {
			t.Errorf("Unexpected eror %v", err)
		}
		out := []byte(tt.out)
		if !bytes.Equal(buf, out) {
			t.Errorf("\nExpected:\n\"%s\",\ngot:\n\"%s\"", tt.out, buf)
		}
	}
}
