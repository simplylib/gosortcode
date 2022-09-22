package cmd

import "bytes"
import "testing"
import _ "embed"

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
