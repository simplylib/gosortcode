// Package cmd handles all the logic for the gosortcode command
package cmd

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/SomethingBot/multierror"
)

func run() error {
	log.SetFlags(log.Flags() | log.Lshortfile)

	//printDiff := flag.Bool("d", false, "print diff")
	writeToFile := flag.Bool("w", false, "write formatted versions back to file, exclusive with -o")
	writeToStdout := flag.Bool("o", false, "write to stdout, exclude with -w")

	flag.CommandLine.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: "+os.Args[0]+" <flags> <filepath>")
		flag.CommandLine.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) != 1 || !*writeToFile && !*writeToStdout {
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	var (
		writer     io.Writer
		sourceFile io.ReadCloser
		err        error
	)
	var f *os.File
	f, err = os.Open(filepath.Clean(flag.Args()[0]))
	if err != nil {
		return fmt.Errorf("could not open file (%v) due to error (%w)", filepath.Clean(flag.Args()[0]), err)
	}
	defer func() { // todo: replace with https://github.com/golang/go/issues/53435
		if err != nil {
			err = multierror.Append(err, f.Close())
		}
	}()

	if *writeToStdout {
		writer = os.Stdout
	} else if *writeToFile {
		var f2 *os.File
		f2, err = os.CreateTemp("", filepath.Base(f.Name()))
		if err != nil {
			return fmt.Errorf("could not create a temporary file (%v)", err)
		}
		defer func() {
			err2 := f.Close()
			if err2 != nil {
				err = multierror.Append(err, err2)
				return
			}
			err2 = os.Remove(f2.Name())
			if err2 != nil {
				err = multierror.Append(err, err2)
			}
			err2 = f2.Close()
			if err2 != nil {
				err = multierror.Append(err, err2)
				return
			}
			err2 = os.Rename(f2.Name(), f.Name())
			if err2 != nil {
				err = multierror.Append(err, err2)
				return
			}
		}()
		writer = f2
	}

	err = format(flag.Args()[0], sourceFile, writer)
	if err != nil {
		return fmt.Errorf("could not parse (%v) due to error (%w)", flag.Args()[0], err)
	}

	return nil
}

// Main function for cmd
func Main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
