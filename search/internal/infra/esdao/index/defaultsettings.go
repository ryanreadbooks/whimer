package index

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelasticanalyzer "github.com/ryanreadbooks/whimer/misc/xelastic/analyzer"
	xnormalizer "github.com/ryanreadbooks/whimer/misc/xelastic/normalizer"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/bulk"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/tokenchar"
)

func defaultSettings(opt *IndexerOption) *types.IndexSettings {
	return &types.IndexSettings{
		MaxNgramDiff: mg.Ptr(5),
		Analysis: &types.IndexSettingsAnalysis{
			// 自定义normalizer和tokenizer和analyzer
			Normalizer: map[string]types.Normalizer{
				"cust_clean_normalizer": xnormalizer.NewCleanNormalizer(),
			},
			Tokenizer: map[string]types.Tokenizer{
				"cust_edge_ngram_tokenizer": &types.EdgeNGramTokenizer{
					MinGram:    mg.Ptr(2),
					MaxGram:    mg.Ptr(7),
					TokenChars: []tokenchar.TokenChar{tokenchar.Letter, tokenchar.Digit},
				},
				"cust_ngram_tokenizer": &types.NGramTokenizer{
					MinGram:    mg.Ptr(2),
					MaxGram:    mg.Ptr(7),
					TokenChars: []tokenchar.TokenChar{tokenchar.Letter, tokenchar.Digit},
				},
			},
			Analyzer: map[string]types.Analyzer{
				"default": xelasticanalyzer.NewIkMaxWordAnalyzer(), // 指定默认analyzer
				"cust_prefix_analyzer": &types.CustomAnalyzer{
					CharFilter: []string{"html_strip"},
					Filter:     []string{"lowercase", "asciifolding", "trim"},
					Tokenizer:  "cust_edge_ngram_tokenizer",
				},
				"cust_ngram_analyzer": &types.CustomAnalyzer{
					CharFilter: []string{"html_strip"},
					Filter:     []string{"lowercase", "asciifolding", "trim"},
					Tokenizer:  "cust_ngram_tokenizer",
				},
			},
		},
		NumberOfReplicas: mg.Ptr(strconv.Itoa(opt.NumberOfReplicas)),
		NumberOfShards:   mg.Ptr(strconv.Itoa(opt.NumbefOfShards)),
	}
}

var (
	defaultTextFields = map[string]types.Property{
		"keyword": &types.KeywordProperty{
			Normalizer: mg.Ptr("cust_clean_normalizer"),
		},
		"prefix": &types.TextProperty{
			Analyzer: mg.Ptr("cust_prefix_analyzer"),
		},
		"ngram": &types.TextProperty{
			Analyzer: mg.Ptr("cust_ngram_analyzer"),
		}}
)

var (
	ErrBulkFailure = xerror.ErrInternal.Msg("es bulk operation err")
)

func handleBulkResponse(ctx context.Context, resp *bulk.Response) error {
	// 一个或者多个错误
	if resp.Errors {
		var errLogs strings.Builder
		errLogs.Grow(256)
		for _, respItem := range resp.Items {
			for k, v := range respItem {
				if v.Error != nil {
					log, _ := v.Error.MarshalJSON()
					errLogs.WriteString(fmt.Sprintf("bulk %s | err: %s", k, log))
				}
			}
		}
		if errLogs.Len() > 0 {
			xlog.Msg(errLogs.String()).Errorx(ctx)
			return ErrBulkFailure
		}
	}

	return nil
}
