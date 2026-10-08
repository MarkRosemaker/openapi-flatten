# Roadmap

Work not yet done. An entry is deleted once it is.

## Webhooks, callbacks and headers

`Document` flattens what paths, operations and components hold, but not
webhooks, callbacks, response headers or the headers of an encoding: their
schemas stay inline, and so does what they contain. A document that describes
its webhooks or callbacks in detail keeps those schemas anonymous downstream,
where a generator has to make up their names.
