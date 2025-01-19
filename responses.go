package flatten

import (
	"github.com/MarkRosemaker/errpath"
	"github.com/MarkRosemaker/openapi"
)

func operationResponses(d *openapi.Document, rs openapi.OperationResponses, opID string) error {
	for code, r := range rs.ByIndex() {
		if err := responseRef(d, r, nameResponse(opID, code), !code.IsSuccess()); err != nil {
			return &errpath.ErrKey{Key: string(code), Err: err}
		}
	}

	return nil
}

func responses(d *openapi.Document, rs openapi.ResponsesByName) error {
	for name, r := range rs.ByIndex() {
		// NOTE: We are *not* calling responseRef here,
		// because we are calling this function from Components,
		// where the response should already be.
		if err := response(d, r.Value, name, isFailureResponse(d, r)); err != nil {
			return &errpath.ErrKey{Key: string(name), Err: err}
		}
	}

	return nil
}

func isFailureResponse(d *openapi.Document, r *openapi.ResponseRef) bool {
	for _, p := range d.Paths {
		for _, o := range p.Operations {
			for code, rs := range o.Responses {
				if rs == r && !code.IsSuccess() {
					return false
				}
			}
		}
	}

	return true
}
