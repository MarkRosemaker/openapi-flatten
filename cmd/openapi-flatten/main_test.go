package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// unsorted has a nested response schema to flatten, its responses out of order, and its component schemas out of
// alphabetical order.
const unsorted = `{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/pets": {
      "get": {
        "operationId": "listPets",
        "responses": {
          "404": {"description": "missing"},
          "200": {"description": "ok", "content": {"application/json": {"schema":
            {"type": "object", "properties": {"name": {"type": "string"}}}}}}
        }
      }
    }
  },
  "components": {"schemas": {
    "Zebra": {"type": "object", "properties": {"stripes": {"type": "integer"}}},
    "Apple": {"type": "object", "properties": {"red": {"type": "boolean"}}}
  }}
}`

func writeSpec(t *testing.T, spec string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "openapi.json")
	if err := os.WriteFile(path, []byte(spec), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func readSpec(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

// inOrder fails unless each of words appears in s, in the order given.
func inOrder(t *testing.T, s string, words ...string) {
	t.Helper()

	last := -1
	for _, w := range words {
		i := strings.Index(s, w)
		if i < 0 {
			t.Fatalf("%s is missing from\n%s", w, s)
		}

		if i < last {
			t.Errorf("%s comes before the word preceding it in %q", w, words)
		}

		last = i
	}
}

// TestRun flattens the spec in place, sorting each operation's responses and the components.
func TestRun(t *testing.T) {
	t.Parallel()

	path := writeSpec(t, unsorted)

	if err := run(t.Context(), []string{"-spec", path}); err != nil {
		t.Fatal(err)
	}

	got := readSpec(t, path)

	if !strings.Contains(got, `"$ref": "#/components/responses/ListPetsOk"`) {
		t.Errorf("the success response is not flattened:\n%s", got)
	}

	inOrder(t, got, `"200"`, `"404"`)
	inOrder(t, got, `"Apple"`, `"ListPetsOk"`, `"Zebra"`)

	if strings.Contains(got, "x-flattened-from") {
		t.Error("marked origins without -mark-origin")
	}
}

func TestRun_MarkOrigin(t *testing.T) {
	t.Parallel()

	path := writeSpec(t, unsorted)

	if err := run(t.Context(), []string{"-spec", path, "-mark-origin"}); err != nil {
		t.Fatal(err)
	}

	if got := readSpec(t, path); !strings.Contains(got,
		`"x-flattened-from": "#/components/responses/ListPetsOk/content/application~1json/schema"`) {
		t.Errorf("the moved schema does not say where it came from:\n%s", got)
	}
}

// A spec flatten cannot handle is left as it was.
func TestRun_FlattenFails(t *testing.T) {
	t.Parallel()

	spec := strings.Replace(unsorted, `"stripes": {"type": "integer"}`, `"stripes": {"type": "foo"}`, 1)
	path := writeSpec(t, spec)

	err := run(t.Context(), []string{"-spec", path})
	if err == nil || !strings.Contains(err.Error(), `components.schemas["Zebra"].properties["stripes"]`) {
		t.Errorf("got %v, want an error naming the stripes property", err)
	}

	if got := readSpec(t, path); got != spec {
		t.Errorf("the spec changed although flattening failed:\n%s", got)
	}
}

func TestRun_Errors(t *testing.T) {
	t.Parallel()

	if err := run(t.Context(), []string{"-spec", filepath.Join(t.TempDir(), "missing.json")}); err == nil {
		t.Error("a spec that does not exist: want an error")
	}

	if err := run(t.Context(), []string{"-spec", writeSpec(t, "{")}); err == nil {
		t.Error("a spec that is not JSON: want an error")
	}

	// after a spec that would flatten, so that only the flag can fail it
	if err := run(t.Context(), []string{"-spec", writeSpec(t, unsorted), "-no-such-flag"}); err == nil {
		t.Error("an unknown flag: want an error")
	}
}

// Without -spec, it flattens api/openapi.json.
func TestRun_DefaultSpec(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "api"), 0o755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "api", "openapi.json")
	if err := os.WriteFile(path, []byte(unsorted), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)

	if err := run(t.Context(), nil); err != nil {
		t.Fatal(err)
	}

	if got := readSpec(t, path); !strings.Contains(got, "ListPetsOk") {
		t.Errorf("api/openapi.json is not flattened:\n%s", got)
	}
}
