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
if err := flatten.Document(doc); err != nil {
    log.Fatal(err)
}

// doc now has no nested inline objects — only $ref pointers
```

`Document` is the entire public API. It modifies the document in place.
