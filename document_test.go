package flatten_test

import (
	_ "embed"
	"testing"

	"github.com/MarkRosemaker/openapi"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

//go:embed openapi.json
var testOpenapi []byte

func TestDocument(t *testing.T) {
	t.Parallel()

	d, err := openapi.LoadFromDataJSON(testOpenapi)
	if err != nil {
		t.Fatal(err)
	}

	if err := flatten.Document(d); err != nil {
		t.Fatal(err)
	}

	s := d.Paths["/organizations/9011051/workspaces"].Post.Responses["403"].Value.Content["application/json"].Schema

	if s.Ref == nil {
		t.Fatal("s.Ref is nil")
	}

	// "/organizations/9011051/workspaces"]["POST"].responses["403"]["application/json"].schema: missing schema ref

}
