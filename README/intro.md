OpenAPI allows schemas to be defined inline anywhere they are used. While convenient
for small specs, deeply nested inline definitions make large specs harder to read,
harder to reuse, and harder to generate consistent client code from. Flattening
gives every meaningful type a name and a single canonical location.

This matters most upstream of code generation: a generator can only emit a named,
reusable Go type for a schema that *has* a name. Flattening first means
[`openapi-codegen`](https://github.com/MarkRosemaker/openapi-codegen) never has to
invent one.
