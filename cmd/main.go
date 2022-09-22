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
	writeToFile := flag.Bool("w", false, "write formatted versions back to file")
	writeToStdout := flag.Bool("o", true, "write to stdout")

	flag.CommandLine.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: "+os.Args[0]+" <flags> <filepath>")
		flag.CommandLine.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	var (
		writer     io.Writer
		sourceFile io.ReadCloser
		err        error
	)
	if *writeToStdout {
		var f *os.File
		f, err = os.Open(filepath.Clean(flag.Args()[0]))
		if err != nil {
			return fmt.Errorf("could not open file (%v) due to error (%w)", filepath.Clean(flag.Args()[0]), err)
		}
		defer func() { // todo: replace with https://github.com/golang/go/issues/53435
			err = multierror.Append(err, f.Close())
		}()
		writer = os.Stdout
	} else if *writeToFile {
		var f *os.File
		f, err = os.OpenFile(filepath.Clean(flag.Args()[0]), os.O_TRUNC|os.O_RDWR, 0o655)
		if err != nil {
			return fmt.Errorf("could not open file (%v) due to error (%w)", filepath.Clean(flag.Args()[0]), err)
		}
		defer func() { // todo: replace with https://github.com/golang/go/issues/53435
			err = multierror.Append(err, f.Close())
		}()

		writer = f
	}

	err = format(flag.Args()[0], sourceFile, writer)
	if err != nil {
		return fmt.Errorf("could not parse (%v) due to error (%w)", flag.Args()[0], err)
	}

	return nil

	return nil
}

func Main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
