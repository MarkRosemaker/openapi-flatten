package flatten_test

import (
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

// bad is a schema flatten does not know what to do with.
const bad = `{"type": "foo"}`

// specWith is a minimal document with paths and components as given, in JSON.
func specWith(paths, components string) string {
	return `{"openapi": "3.1.0", "info": {"title": "t", "version": "1"}, "paths": {` + paths +
		`}, "components": {` + components + `}}`
}

// operationWith is the paths of a single GET /pets, with the given fields besides its responses.
func operationWith(fields, responses string) string {
	if fields != "" {
		fields += ","
	}

	return `"/pets": {"get": {"operationId": "listPets", ` + fields + ` "responses": {` + responses + `}}}`
}

const okResponse = `"200": {"description": "ok"}`

// TestDocument_ErrorPath: an error says where in the document it is, as a path down to the field, for every part of
// the document Document walks.
func TestDocument_ErrorPath(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, spec, want string
	}{
		{
			"path parameter",
			specWith(`"/pets/{id}": {"parameters": [{"name": "id", "in": "path", "required": true, "schema": `+bad+`}],
				"get": {"operationId": "getPet", "responses": {`+okResponse+`}}}`, ""),
			`paths["/pets/{id}"].parameters[0].schema: unimplemented schema ref type "foo"`,
		},
		{
			"operation parameter content",
			specWith(operationWith(`"parameters": [{"name": "q", "in": "query",
				"content": {"application/json": {"schema": `+bad+`}}}]`, okResponse), ""),
			`paths["/pets"].GET.parameters[0].content["application/json"].schema: unimplemented schema ref type "foo"`,
		},
		{
			"request body",
			specWith(operationWith(`"requestBody": {"content": {"application/json": {"schema": `+bad+`}}}`,
				okResponse), ""),
			`paths["/pets"].GET.requestBody.content["application/json"].schema: unimplemented schema ref type "foo"`,
		},
		{
			"property of a success response",
			specWith(operationWith("", `"200": {"description": "ok", "content": {"application/json": {"schema":
				{"type": "object", "properties": {"a": `+bad+`}}}}}`), ""),
			`paths["/pets"].GET.responses["200"].content["application/json"].schema.properties["a"]: ` +
				`unimplemented schema ref type "foo"`,
		},
		{
			// a failure's schema is always moved, so it is the moved schema that is refused
			"failure response",
			specWith(operationWith("", `"404": {"description": "missing", "content": {"application/json":
				{"schema": `+bad+`}}}`), ""),
			`paths["/pets"].GET.responses["404"].content["application/json"].schema: unimplemented schema type "foo"`,
		},
		{
			"array items of an unknown type",
			specWith("", `"schemas": {"Pet": {"type": "object", "properties": {"tags": {"type": "array",
				"items": `+bad+`}}}}`),
			`components.schemas["Pet"].properties["tags"]: unimplemented item type "foo"`,
		},
		{
			"component schema",
			specWith("", `"schemas": {"Pet": `+bad+`}`),
			`components.schemas["Pet"]: unimplemented schema type "foo"`,
		},
		{
			"allOf entry",
			specWith("", `"schemas": {"Pet": {"allOf": [{"required": ["id"]}, `+bad+`]}}`),
			`components.schemas["Pet"].allOf[1]: unimplemented schema ref type "foo"`,
		},
		{
			"oneOf alternative",
			specWith("", `"schemas": {"Pet": {"oneOf": [`+bad+`]}}`),
			`components.schemas["Pet"].oneOf[0]: unimplemented schema ref type "foo"`,
		},
		{
			"anyOf alternative",
			specWith("", `"schemas": {"Pet": {"anyOf": [{"type": "string"}, `+bad+`]}}`),
			`components.schemas["Pet"].anyOf[1]: unimplemented schema ref type "foo"`,
		},
		{
			"tuple position",
			specWith("", `"schemas": {"Pair": {"type": "array", "prefixItems": [`+bad+`]}}`),
			`components.schemas["Pair"].prefixItems[0]: unimplemented schema ref type "foo"`,
		},
		{
			"items",
			specWith("", `"schemas": {"Pets": {"type": "array", "items": `+bad+`}}`),
			`components.schemas["Pets"].items: unimplemented schema ref type "foo"`,
		},
		{
			"additionalProperties",
			specWith("", `"schemas": {"Map": {"type": "object", "additionalProperties": `+bad+`}}`),
			`components.schemas["Map"].additionalProperties: unimplemented schema ref type "foo"`,
		},
		{
			"propertyNames",
			specWith("", `"schemas": {"Map": {"type": "object", "propertyNames": `+bad+`}}`),
			`components.schemas["Map"].propertyNames: unimplemented schema ref type "foo"`,
		},
		{
			"not",
			specWith("", `"schemas": {"Map": {"type": "object", "not": `+bad+`}}`),
			`components.schemas["Map"].not: unimplemented schema ref type "foo"`,
		},
		{
			"component response",
			specWith("", `"responses": {"NotFound": {"description": "missing", "content": {"application/json":
				{"schema": `+bad+`}}}}`),
			`components.responses["NotFound"].content["application/json"].schema: unimplemented schema ref type "foo"`,
		},
		{
			"component parameter",
			specWith("", `"parameters": {"Limit": {"name": "limit", "in": "query", "schema": `+bad+`}}`),
			`components.parameters["Limit"].schema: unimplemented schema ref type "foo"`,
		},
		{
			"component request body",
			specWith("", `"requestBodies": {"NewPet": {"content": {"application/json": {"schema": `+bad+`}}}}`),
			`components.requestBodies["NewPet"].content["application/json"].schema: unimplemented schema ref type "foo"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := openapi.LoadFromDataJSON([]byte(tc.spec))
			if err != nil {
				t.Fatal(err)
			}

			err = flatten.Document(doc, flatten.Config{})
			if err == nil {
				t.Fatalf("want an error at %s", strings.SplitN(tc.want, ":", 2)[0])
			}

			if err.Error() != tc.want {
				t.Errorf("got  %s\nwant %s", err, tc.want)
			}
		})
	}
}
