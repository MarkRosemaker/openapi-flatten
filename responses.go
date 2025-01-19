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

func responses(d *openapi.Document, rs openapi.ResponsesByName, rspName string) error {
	for name, r := range rs.ByIndex() {
		if err := responseRef(d, r, rspName, false); err != nil {
			return &errpath.ErrKey{Key: string(name), Err: err}
		}
	}

	return nil
}
