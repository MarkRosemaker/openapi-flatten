# Roadmap

- **Visit `propertyNames`.** MarkRosemaker/openapi#22 added `Schema.PropertyNames`, a schema for an object's keys. Once `go.mod` is on a version that has it, `schema.go` should flatten it like `additionalProperties`, so an inline key enum (pixellab's compass directions) becomes a named component.
