package flatten_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

// TestDocument_Golden flattens testdata/before.json, marking origins, three times over, and must get
// testdata/after.json each time: flattening a flat document changes nothing. Both are edited by hand; what each part
// of before.json is there for is said in it.
func TestDocument_Golden(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromFile("testdata/before.json")
	if err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile("testdata/after.json")
	if err != nil {
		t.Fatal(err)
	}

	for run := range 3 {
		if err := flatten.Document(doc, flatten.Config{MarkOrigin: true}); err != nil {
			t.Fatalf("run %d: %v", run+1, err)
		}

		if err := doc.Validate(); err != nil {
			t.Fatalf("run %d: %v", run+1, err)
		}

		got, err := doc.ToJSON()
		if err != nil {
			t.Fatal(err)
		}

		got = append(got, '\n')

		if line, ok := firstDifference(got, want); !ok {
			t.Fatalf("run %d: after.json %s", run+1, line)
		}
	}
}

// firstDifference describes the first line in which got and want differ, if any.
func firstDifference(got, want []byte) (string, bool) {
	if bytes.Equal(got, want) {
		return "", true
	}

	gotLines, wantLines := bytes.Split(got, []byte("\n")), bytes.Split(want, []byte("\n"))
	for i := range min(len(gotLines), len(wantLines)) {
		if !bytes.Equal(gotLines[i], wantLines[i]) {
			return fmt.Sprintf("line %d: got %s, want %s", i+1, bytes.TrimSpace(gotLines[i]), bytes.TrimSpace(wantLines[i])), false
		}
	}

	return fmt.Sprintf("has %d lines, got %d", len(wantLines), len(gotLines)), false
}

// TestDocument_MarkOriginOff: without MarkOrigin, no schema says where it came from.
func TestDocument_MarkOriginOff(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromFile("testdata/before.json")
	if err != nil {
		t.Fatal(err)
	}

	if err := flatten.Document(doc, flatten.Config{}); err != nil {
		t.Fatal(err)
	}

	for name, s := range doc.Components.Schemas {
		if bytes.Contains(s.Extensions, []byte(flatten.ExtensionOrigin)) {
			t.Errorf("%s is marked without MarkOrigin", name)
		}
	}
}

// TestDocument_OnePathKeepsItsPrefix: a prefix is moved into the servers only when several paths share it, as one path
// alone says nothing of what is common to the API.
func TestDocument_OnePathKeepsItsPrefix(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "servers": [{"url": "https://api.example.com"}],
  "paths": {"/v1/pets": {"get": {"responses": {"200": {"description": "ok"}}}}}
}`))
	if err != nil {
		t.Fatal(err)
	}

	if err := flatten.Document(doc, flatten.Config{}); err != nil {
		t.Fatal(err)
	}

	if _, ok := doc.Paths["/v1/pets"]; !ok || doc.Servers[0].URL != "https://api.example.com" {
		t.Errorf("got paths %v and server %s, want them as they were", doc.Paths, doc.Servers[0].URL)
	}
}
