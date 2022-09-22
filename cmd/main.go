package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/SomethingBot/multierror"
)

func run() error {
	log.SetFlags(log.Flags() | log.Lshortfile)

	printDiff := flag.Bool("d", false, "print diff")
	writeToFile := flag.Bool("w", false, "write formatted versions back to file")
	writeToStdout := flag.Bool("o", true, "write to stdout")

	flag.CommandLine.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: "+os.Args[0]+" <flags> <filepath>")
		flag.CommandLine.PrintDefaults()
	}

	flag.Parse()

	_ = printDiff
	_ = writeToFile
	_ = writeToStdout

	if len(flag.Args()) != 1 {
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	sourceFile, err := os.Open(filepath.Clean(flag.Args()[0]))
	if err != nil {
		return fmt.Errorf("could not open file (%v) due to error (%w)", filepath.Clean(flag.Args()[0]), err)
	}
	defer func() { // todo: replace with https://github.com/golang/go/issues/53435
		err = multierror.Append(err, sourceFile.Close())
	}()

	err = format(sourceFile.Name(), sourceFile, os.Stdout)
	if err != nil {
		return fmt.Errorf("could not parse (%v) due to error (%w)", sourceFile.Name(), err)
	}

	return nil
}

func Main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
