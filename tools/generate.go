package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
	"github.com/MarkRosemaker/fsutil/osutil"
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

		if err := doc.WriteToFile(filepath.Join("testdata", entry.Name(), "golden.json")); err != nil {
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
		if err := osutil.Copy(
			filepath.Join(srcDir, e.Name(), "golden.json"),
			filepath.Join("testdata", e.Name(), "openapi.json"),
		); err != nil {
			return err
		}
	}

	return nil
}
