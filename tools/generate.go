package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if err := copyPreviousStep(); err != nil {
		return err
	}

	entries, err := os.ReadDir("testdata")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		doc, err := openapi.LoadFromFile(filepath.Join("testdata", entry.Name(), "openapi.json"))
		if err != nil {
			return err
		}

		if err := flatten.Document(doc); err != nil {
			return err
		}

		if err := writeJSON(filepath.Join("testdata", entry.Name(), "golden.json"), doc); err != nil {
			return err
		}
	}

	return nil
}

func copyPreviousStep() error {
	const srcDir = "../openapi-enrich/testdata"

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return err
	}

	for _, e := range entries {
		if err := copyFile(
			filepath.Join(srcDir, e.Name(), "golden.json"),
			filepath.Join("testdata", e.Name(), "openapi.json"),
		); err != nil {
			return err
		}
	}

	return nil
}

func writeJSON(path string, doc *openapi.Document) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := doc.WriteJSON(f); err != nil {
		return fmt.Errorf("writing to %q: %w", path, err)
	}

	return nil
}

func copyFile(srcName, dstName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
