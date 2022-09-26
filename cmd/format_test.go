package cmd

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	_ "embed"

	"github.com/dave/dst/decorator"
)

//go:embed testdata/unsorted.go
var unsortedFile []byte

//go:embed testdata/sorted.go
var sortedFile []byte

func TestFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	err := format("testdata/unsorted.go", bytes.NewReader(unsortedFile), buf)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(sortedFile, buf.Bytes()) != 0 {
		t.Fatalf("wanted:\n%v\ngot:\n%v\n", string(sortedFile), string(buf.Bytes()))
	}
}

func BenchmarkSortASTFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		astFile, err := decorator.ParseFile(token.NewFileSet(), "testdata/unsorted.go", bytes.NewReader(unsortedFile), parser.ParseComments)
		if err != nil {
			b.Fatal(fmt.Errorf("could not parse ast (%w)", err))
		}
		err = sortASTFile(astFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}
