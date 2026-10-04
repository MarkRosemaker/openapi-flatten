package flatten_test

import (
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

func TestFlatten_ComponentResponseSchemas(t *testing.T) {
	// a component response promotes even a simple schema only if an operation uses it for a failure
	doc, err := openapi.LoadFromReader(strings.NewReader(`{
		"openapi": "3.1.0",
		"info": {"title": "test", "version": "0.0.1"},
		"paths": {"/pets": {"get": {"operationId": "listPets", "responses": {
			"200": {"$ref": "#/components/responses/Names"},
			"404": {"$ref": "#/components/responses/Missing"}
		}}}},
		"components": {"responses": {
			"Names": {"description": "ok", "content": {"application/json": {"schema": {"type": "array", "items": {"type": "string"}}}}},
			"Missing": {"description": "not found", "content": {"application/json": {"schema": {"type": "array", "items": {"type": "string"}}}}}
		}}
	}`))
	if err != nil {
		t.Fatal(err)
	}

	if err := flatten.Document(doc, flatten.Config{}); err != nil {
		t.Fatal(err)
	}

	schemaOf := func(name string) *openapi.Schema {
		return doc.Components.Responses[name].Value.Content[openapi.MediaRangeJSON].Schema
	}

	if s := schemaOf("Names"); s.Ref != nil {
		t.Errorf("Names: got a reference to %s, want the array inline", s.Ref.Identifier)
	}

	if s := schemaOf("Missing"); s.Ref == nil || s.Ref.Identifier != "#/components/schemas/Missing" {
		t.Errorf("Missing: got %+v, want a reference to the Missing schema", s)
	}
}
