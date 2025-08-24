package normalizer

import "github.com/elastic/go-elasticsearch/v8/typedapi/types"

func NewCleanNormalizer() types.Normalizer {
	return &types.CustomNormalizer{
		Filter: []string{"lowercase", "asciifolding", "trim"},
	}
}