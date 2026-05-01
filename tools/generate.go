package main

import (
	"context"
	"fmt"
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

		doc.SortMaps()

		if err := writeJSON(filepath.Join("testdata", entry.Name(), "golden.json"), doc); err != nil {
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
