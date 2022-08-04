package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

func main() {
	if err := run(); err != nil {
		log.New(os.Stderr, "error: ", 0).Fatal(err)
	}
}

func run() error {
	inputFileFlag := flag.String("i", "input.swf", "specify the swf file name for input")
	outputFileFlag := flag.String("o", "output.swf", "specify the swf file name for output")
	versionFlag := flag.Bool("v", false, "print version")

	flag.Parse()

	if *versionFlag {
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(info.Main.Version)
		} else {
			fmt.Println("undefined")
		}

		return nil
	}

	input, err := os.Open(*inputFileFlag)

	if err != nil {
		return err
	}

	defer input.Close()

	header := &bytes.Buffer{}

	written, err := io.CopyN(header, input, 8)

	if err != nil {
		return err
	}
	if written != 8 {
		return fmt.Errorf("something went wrong while parsing swf file header")
	}

	signature := &bytes.Buffer{}

	if _, err := io.CopyN(signature, header, 3); err != nil {
		return err
	}
	if string(signature.Bytes()) == `FWS` {
		return fmt.Errorf("%s is already uncompressed", *inputFileFlag)
	}
	if string(signature.Bytes()) != `CWS` {
		return fmt.Errorf("not a compressed swf file")
	}

	content, err := zlib.NewReader(input)

	if err != nil {
		return err
	}

	defer content.Close()

	temporaryOutputPath := filepath.Join(os.TempDir(), filepath.Base(*outputFileFlag))
	actualOutputPath := *outputFileFlag

	output, err := os.Create(temporaryOutputPath)

	if err != nil {
		return err
	}

	defer output.Close()

	// The header variable contains 5 bytes, which consists of 1 byte for Macromedia Flash version and 4 bytes for file size.
	uncompressedHeader := bytes.NewBuffer([]byte(`FWS`))

	if _, err := io.Copy(uncompressedHeader, header); err != nil {
		return fmt.Errorf("failed to build uncompressed header: %w")
	}
	if _, err := io.Copy(output, uncompressedHeader); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := io.Copy(output, content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}
	if err := os.Rename(temporaryOutputPath, actualOutputPath); err != nil {
		return err
	}

	return nil
}
