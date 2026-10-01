# Roadmap

Work not yet done. An entry is deleted once it is.

## A success response in components promotes every schema

A response in `components/responses` should promote its schema only if the
schema is complex, unless an operation uses the response for a status other
than a success. `isFailureResponse` in `responses.go` gets this backwards on
two counts:

- it returns `false` when it finds the response under a non-success status,
  and `true` otherwise;
- it compares an operation's `*ResponseRef` with the component's by pointer,
  and once the operation holds a `$ref` the two are never the same, so it
  never finds one.

So every component response promotes its schema. A success-only response
whose schema is a plain array of strings still gets that array named.
Responses written inline in an operation are not affected.
