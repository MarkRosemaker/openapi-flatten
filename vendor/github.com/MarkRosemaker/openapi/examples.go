package openapi

import (
	"iter"

	"github.com/MarkRosemaker/errpath"
	"github.com/MarkRosemaker/ordmap"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// Examples is a map of examples.
type Examples map[string]*ExampleRef

// Validate validates the map of examples.
func (exs Examples) Validate() error {
	for k, ex := range exs.ByIndex() {
		if err := validateKey(k); err != nil {
			return err
		}

		if err := ex.Validate(); err != nil {
			return &errpath.ErrKey{Key: k, Err: err}
		}
	}

	return nil
}

// ByIndex returns a sequence of key-value pairs ordered by index.
func (exs Examples) ByIndex() iter.Seq2[string, *ExampleRef] {
	return ordmap.ByIndex(exs, getIndexRef[Example, *Example])
}

// Sort sorts the map by key and sets the indices accordingly.
func (exs Examples) Sort() {
	ordmap.Sort(exs, setIndexRef[Example, *Example])
}

// Set sets a value in the map, adding it at the end of the order.
func (exs *Examples) Set(key string, ex *ExampleRef) {
	ordmap.Set(exs, key, ex, getIndexRef[Example, *Example], setIndexRef[Example, *Example])
}

// MarshalJSONTo marshals the key-value pairs in order.
func (exs *Examples) MarshalJSONTo(enc *jsontext.Encoder, opts json.Options) error {
	return ordmap.MarshalJSONTo(exs, enc, opts)
}

// UnmarshalJSONFrom unmarshals the key-value pairs in order and sets the indices.
func (exs *Examples) UnmarshalJSONFrom(dec *jsontext.Decoder, opts json.Options) error {
	return ordmap.UnmarshalJSONFrom(exs, dec, opts, setIndexRef[Example, *Example])
}

func (l *loader) collectExamples(exs Examples, ref ref) {
	for k, ex := range exs.ByIndex() {
		l.collectExampleRef(ex, append(ref, k))
	}
}

func (l *loader) resolveExamples(exs Examples) error {
	for k, ex := range exs.ByIndex() {
		if err := l.resolveExampleRef(ex); err != nil {
			return &errpath.ErrKey{Key: k, Err: err}
		}
	}

	return nil
}
