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
