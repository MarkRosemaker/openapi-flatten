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
