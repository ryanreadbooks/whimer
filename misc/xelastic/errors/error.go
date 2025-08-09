package errors

import (
	stderr "errors"
	
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func IsElasticError(err error) bool {
	_, ok := err.(*types.ElasticsearchError)
	return ok
}

func IsResourceAlreadyExistsError(err error) bool {
	var e *types.ElasticsearchError
	if stderr.As(err, &e) {
		return e.ErrorCause.Type == ResourceAlreadyExistsException
	}
	return false
}
