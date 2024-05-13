package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		if len(os.Args) != 4 || os.Args[2] != "-p" {
			os.Exit(1)
		}
		blobSHA := os.Args[3]
		_, err := os.Stat(".git")
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Not a git repository\n")
			os.Exit(1)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading .git directory: %s\n", err)
			os.Exit(1)
		}

		data, err := readBlobData(blobSHA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading blob: %s\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%s", string(data))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func readBlobData(blobSHA string) ([]byte, error) {
	filePath := fmt.Sprintf(".git/objects/%s/%s", blobSHA[:2], blobSHA[2:])
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	bReader := bytes.NewReader(fileBytes)

	reader, err := zlib.NewReader(bReader)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	unzipped, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	index := bytes.IndexByte(unzipped, 0)
	if index == -1 {
		return nil, fmt.Errorf("Blob corrupted")
	}
	return unzipped[index+1:], nil
}
