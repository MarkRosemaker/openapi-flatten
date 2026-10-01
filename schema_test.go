package flatten_test

import (
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

// minimalDoc wraps a schema JSON fragment into the smallest valid OpenAPI 3.1
// document that exercises schemaRef via a GET /test response body.
func minimalDoc(t *testing.T, schemaJSON string) *openapi.Document {
	t.Helper()

	raw := `{
		"openapi": "3.1.0",
		"info": {"title": "test", "version": "0.0.1"},
		"paths": {
			"/test": {
				"get": {
					"responses": {
						"200": {
							"description": "ok",
							"content": {
								"application/json": {
									"schema": ` + schemaJSON + `
								}
							}
						}
					}
				}
			}
		}
	}`

	doc, err := openapi.LoadFromReader(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("LoadFromReader: %v", err)
	}

	return doc
}

func TestFlatten_TypeNull(t *testing.T) {
	doc := minimalDoc(t, `{"type": "null"}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatalf("unexpected error for TypeNull schema: %v", err)
	}
}

func TestFlatten_ArrayOfBoolean(t *testing.T) {
	doc := minimalDoc(t, `{"type": "array", "items": {"type": "boolean"}}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatalf("unexpected error for array-of-boolean schema: %v", err)
	}
}

func TestFlatten_ArrayOfNull(t *testing.T) {
	doc := minimalDoc(t, `{"type": "array", "items": {"type": "null"}}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatalf("unexpected error for array-of-null schema: %v", err)
	}
}

func TestFlatten_TupleArray(t *testing.T) {
	// prefixItems only, no items: a tuple whose length is fully known.
	doc := minimalDoc(t, `{
		"type": "array",
		"prefixItems": [
			{"type": "string"},
			{"type": "integer"}
		]
	}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatalf("unexpected error for tuple-shaped array schema: %v", err)
	}

	if len(doc.Components.Schemas) != 1 {
		t.Fatalf("Components.Schemas: got %d entries, want 1 (the tuple itself)", len(doc.Components.Schemas))
	}
}

func TestFlatten_TupleArray_ObjectPosition(t *testing.T) {
	// a prefixItems entry with properties needs a name of its own, the same
	// as any other object schema reached through items or properties.
	doc := minimalDoc(t, `{
		"type": "array",
		"prefixItems": [
			{"type": "string"},
			{"type": "object", "properties": {"id": {"type": "integer"}}}
		]
	}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatalf("unexpected error for tuple-with-object-position schema: %v", err)
	}

	if len(doc.Components.Schemas) != 2 {
		t.Fatalf("Components.Schemas: got %d entries, want 2 (the tuple and its object position)", len(doc.Components.Schemas))
	}
}

func TestFlatten_ArrayOfAnyOf(t *testing.T) {
	// items with no explicit type, using anyOf (e.g. nullable union)
	doc := minimalDoc(t, `{
		"type": "array",
		"items": {
			"anyOf": [
				{"type": "string", "format": "date-time"},
				{"type": "null"}
			]
		}
	}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatalf("unexpected error for array-of-anyOf schema: %v", err)
	}
}

func TestFlatten_UnionBranches(t *testing.T) {
	// a branch with a shape of its own gets a name: its title if it has one, else parent, keyword and position
	doc := minimalDoc(t, `{
		"oneOf": [
			{"title": "A person", "type": "object", "properties": {"name": {"type": "string"}}},
			{"type": "object", "properties": {"id": {"type": "integer"}}},
			{"type": "string"},
			{"type": "null"}
		]
	}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatal(err)
	}

	var names []string
	for name := range doc.Components.Schemas.ByIndex() {
		names = append(names, name)
	}

	if got, want := strings.Join(names, " "), "Ok APerson OkOneOf1"; got != want {
		t.Fatalf("got components %s, want %s", got, want)
	}

	union := doc.Components.Schemas["Ok"]
	for i, want := range []string{"#/components/schemas/APerson", "#/components/schemas/OkOneOf1", "", ""} {
		got := ""
		if r := union.OneOf[i].Ref; r != nil {
			got = r.Identifier
		}

		if got != want {
			t.Errorf("oneOf[%d]: got ref %q, want %q", i, got, want)
		}
	}
}

func TestFlatten_PropertyNamesAndNot(t *testing.T) {
	// the schema for an object's keys and the schema it must not match are flattened like any other
	doc := minimalDoc(t, `{
		"type": "object",
		"properties": {"id": {"type": "integer"}},
		"propertyNames": {"type": "string", "enum": ["id", "north"]},
		"not": {"type": "object", "properties": {"deleted": {"type": "boolean"}}}
	}`)

	if err := flatten.Document(doc); err != nil {
		t.Fatal(err)
	}

	var names []string
	for name := range doc.Components.Schemas.ByIndex() {
		names = append(names, name)
	}

	if got, want := strings.Join(names, " "), "Ok OkKey OkNot"; got != want {
		t.Fatalf("got components %s, want %s", got, want)
	}

	ok := doc.Components.Schemas["Ok"]
	if r := ok.PropertyNames.Ref; r == nil || r.Identifier != "#/components/schemas/OkKey" {
		t.Errorf("propertyNames: got %+v, want a reference to OkKey", ok.PropertyNames)
	}

	if r := ok.Not.Ref; r == nil || r.Identifier != "#/components/schemas/OkNot" {
		t.Errorf("not: got %+v, want a reference to OkNot", ok.Not)
	}
}

func TestFlatten_AllOfEntries(t *testing.T) {
	// an allOf entry with a shape of its own is named like any other schema
	doc := minimalDoc(t, `{
		"allOf": [
			{"type": "object", "properties": {"id": {"type": "string"}}},
			{"oneOf": [
				{"title": "A cat", "type": "object", "properties": {"meow": {"type": "boolean"}}},
				{"type": "object", "properties": {"bark": {"type": "boolean"}}}
			]},
			{"required": ["id"]}
		]
	}`)

	for range 3 {
		if err := flatten.Document(doc); err != nil {
			t.Fatal(err)
		}
	}

	var names []string
	for name := range doc.Components.Schemas.ByIndex() {
		names = append(names, name)
	}

	if got, want := strings.Join(names, " "), "Ok OkAllOf0 OkAllOf1 ACat OkAllOf1OneOf1"; got != want {
		t.Fatalf("got components %s, want %s", got, want)
	}

	ok := doc.Components.Schemas["Ok"]
	for i, want := range []string{"#/components/schemas/OkAllOf0", "#/components/schemas/OkAllOf1", ""} {
		got := ""
		if r := ok.AllOf[i].Ref; r != nil {
			got = r.Identifier
		}

		if got != want {
			t.Errorf("allOf[%d]: got ref %q, want %q", i, got, want)
		}
	}

	if err := doc.Validate(); err != nil {
		t.Fatal(err)
	}
}
