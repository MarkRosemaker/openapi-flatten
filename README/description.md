---
tagline: Give every type in your API spec a name and a home.
logo:
    alt: A gopher flattening a stack of nested spec pages with a rolling pin, beside a tidy row of separate named cards
    source: openapi-flatten.jpg
    width: 500
---

`openapi-flatten` eliminates nesting in [OpenAPI 3.x](https://spec.openapis.org/oas/v3.1.0)
specifications. It promotes inline schema definitions, responses, request bodies,
and parameters into the top-level `components` section and replaces them with `$ref`
references.
