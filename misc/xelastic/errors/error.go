package errors

import (
	stderr "errors"
	
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func IsElastic(err error) bool {
	_, ok := err.(*types.ElasticsearchError)
	return ok
}

func IsResourceAlreadyExists(err error) bool {
	var e *types.ElasticsearchError
	if stderr.As(err, &e) {
		return e.ErrorCause.Type == ResourceAlreadyExistsException
	}
	return false
}
