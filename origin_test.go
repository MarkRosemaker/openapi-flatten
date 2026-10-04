package flatten_test

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"testing"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

const originSpec = `{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/pages/{id}": {"get": {
      "operationId": "getPage",
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "responses": {"200": {"description": "ok", "content": {"application/json": {"schema": {
        "type": "object",
        "properties": {"cover": {"type": "object", "properties": {"url": {"type": "string"}}}}
      }}}}}
    }}
  },
  "components": {"schemas": {
    "Page": {"type": "object", "x-go-name": "Page", "properties": {
      "parent": {"type": "object", "properties": {"id": {"type": "string"}}}
    }}
  }}
}`

func origin(t *testing.T, s *openapi.Schema) string {
	t.Helper()

	var ext map[string]jsontext.Value

	if len(s.Extensions) == 0 {
		return ""
	}

	if err := json.Unmarshal(s.Extensions, &ext); err != nil {
		t.Fatal(err)
	}

	var from string
	if v, ok := ext[flatten.ExtensionOrigin]; ok {
		if err := json.Unmarshal(v, &from); err != nil {
			t.Fatal(err)
		}
	}

	return from
}

func TestDocument_MarkOrigin(t *testing.T) {
	doc, err := openapi.LoadFromDataJSON([]byte(originSpec))
	if err != nil {
		t.Fatal(err)
	}

	if err := flatten.Document(doc, flatten.Config{MarkOrigin: true}); err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for name, s := range doc.Components.Schemas {
		got[name] = origin(t, s)
	}

	// a schema flatten moved says where its reference is; one that was a component already says nothing
	if got["Page"] != "" {
		t.Errorf("Page was a component already, but is marked from %q", got["Page"])
	}

	want := map[string]bool{
		"#/components/schemas/Page/properties/parent":                       false,
		"#/components/responses/GetPageOk/content/application~1json/schema": false,
		"#/components/schemas/GetPageOk/properties/cover":                   false,
	}

	for name, from := range got {
		if name == "Page" {
			continue
		}

		if from == "" {
			t.Errorf("%s was moved but is not marked", name)
		}

		if _, ok := want[from]; ok {
			want[from] = true
		}
	}

	for from, seen := range want {
		if !seen {
			t.Errorf("no schema is marked from %s; got %v", from, got)
		}
	}

	// the extensions a schema had are kept
	if string(doc.Components.Schemas["Page"].Extensions) != `{"x-go-name":"Page"}` {
		t.Errorf("Page's extensions are %s", doc.Components.Schemas["Page"].Extensions)
	}

	if err := doc.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDocument_MarkOriginOff(t *testing.T) {
	doc, err := openapi.LoadFromDataJSON([]byte(originSpec))
	if err != nil {
		t.Fatal(err)
	}

	if err := flatten.Document(doc, flatten.Config{}); err != nil {
		t.Fatal(err)
	}

	for name, s := range doc.Components.Schemas {
		if from := origin(t, s); from != "" {
			t.Errorf("%s is marked from %q without MarkOrigin", name, from)
		}
	}
}
