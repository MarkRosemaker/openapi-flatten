```bash
go get -tool github.com/MarkRosemaker/openapi-flatten/cmd/openapi-flatten
```

or

```bash
go get github.com/MarkRosemaker/openapi-flatten
```


```go
import (
    "github.com/MarkRosemaker/openapi"
    flatten "github.com/MarkRosemaker/openapi-flatten"
)

// Load an OpenAPI document (JSON or YAML)
doc, err := openapi.LoadFromDataJSON(jsonBytes)
if err != nil {
    log.Fatal(err)
}

// Flatten all inline definitions
if err := flatten.Document(doc, flatten.Config{}); err != nil {
    log.Fatal(err)
}

// doc now has no nested inline objects — only $ref pointers
```

`Document` is the entire public API. It modifies the document in place.

With `Config{MarkOrigin: true}`, each schema `Document` moves into
`components/schemas` gets an `x-flattened-from` extension: a JSON pointer to the
`$ref` left in its place, such as `#/components/schemas/Page/properties/cover`.
A schema without it was a component already, so a later step can tell a name the
specification gave from one flatten made up. `openapi-compress` reads it, to
prefer the specification's names when it merges schemas, and removes it. The
command line tool takes `-mark-origin` for it.
