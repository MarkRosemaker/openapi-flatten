package flatten

import (
	"github.com/MarkRosemaker/openapi"
)

func parameterRef(d *openapi.Document, p *openapi.ParameterRef) error {
	if p.Ref != nil {
		return nil
	}

	// reference the parameter in the components
	paramName := uniqueName(d.Components.Parameters, p.Value.Name)
	d.Components.Parameters.Set(paramName, &openapi.ParameterRef{Value: p.Value})
	p.Ref = newRef("parameters", paramName)

	return parameter(d, p.Value)
}

func parameter(d *openapi.Document, p *openapi.Parameter) error {
	// if p.Schema != nil {
	// 	if err := l.resolveSchema(p.Schema); err != nil {
	// 		return &errpath.ErrField{Field: "schema", Err: err}
	// 	}
	// }

	// if err := l.resolveContent(p.Content); err != nil {
	// 	return &errpath.ErrField{Field: "content", Err: err}
	// }

	// if err := l.resolveExamples(p.Examples); err != nil {
	// 	return &errpath.ErrField{Field: "examples", Err: err}
	// }

	return nil
}
