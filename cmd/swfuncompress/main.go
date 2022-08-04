package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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
	if string(header.Bytes()[:3]) == `FWS` {
		return fmt.Errorf("%s is already uncompressed", *inputFileFlag)
	}
	if string(header.Bytes()[:3]) != `CWS` {
		return fmt.Errorf("not a compressed swf file")
	}

	reader, err := zlib.NewReader(input)

	if err != nil {
		return err
	}

	defer reader.Close()

	output, err := os.Create(*outputFileFlag)

	if err != nil {
		return err
	}

	defer output.Close()

	uncompressedHeader := bytes.NewBuffer([]byte(`FWS`))

	if _, err := io.Copy(uncompressedHeader, bytes.NewBuffer(header.Bytes()[3:])); err != nil {
		return fmt.Errorf("failed to build header: %w")
	}
	if _, err := io.Copy(output, uncompressedHeader); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := io.Copy(output, reader); err != nil {
		return fmt.Errorf("failed to write content: %w")
	}

	return nil
}
