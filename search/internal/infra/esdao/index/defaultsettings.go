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

const (
	CustomCleanNormalizer    = "cust_clean_normalizer"
	CustomEdgeNgramTokenizer = "cust_edge_ngram_tokenizer"
	CustomNgramTokenizer     = "cust_ngram_tokenizer"
	CustomPrefixAnalyzer     = "cust_prefix_analyzer"
	CustomNgramAnalyzer      = "cust_ngram_analyzer"
)

func defaultSettings(opt *IndexerOption) *types.IndexSettings {
	return &types.IndexSettings{
		MaxNgramDiff: mg.Ptr(5),
		Analysis: &types.IndexSettingsAnalysis{
			// 自定义normalizer和tokenizer和analyzer
			Normalizer: map[string]types.Normalizer{
				CustomCleanNormalizer: xnormalizer.NewCleanNormalizer(),
			},
			Tokenizer: map[string]types.Tokenizer{
				CustomEdgeNgramTokenizer: &types.EdgeNGramTokenizer{
					MinGram:    mg.Ptr(2),
					MaxGram:    mg.Ptr(7),
					TokenChars: []tokenchar.TokenChar{tokenchar.Letter, tokenchar.Digit},
				},
				CustomNgramTokenizer: &types.NGramTokenizer{
					MinGram:    mg.Ptr(2),
					MaxGram:    mg.Ptr(7),
					TokenChars: []tokenchar.TokenChar{tokenchar.Letter, tokenchar.Digit},
				},
			},
			Analyzer: map[string]types.Analyzer{
				"default": xelasticanalyzer.NewIkMaxWordAnalyzer(), // 指定默认analyzer
				CustomPrefixAnalyzer: &types.CustomAnalyzer{
					CharFilter: []string{"html_strip"},
					Filter:     []string{"lowercase", "asciifolding", "trim"},
					Tokenizer:  CustomEdgeNgramTokenizer,
				},
				CustomNgramAnalyzer: &types.CustomAnalyzer{
					CharFilter: []string{"html_strip"},
					Filter:     []string{"lowercase", "asciifolding", "trim"},
					Tokenizer:  CustomNgramTokenizer,
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
