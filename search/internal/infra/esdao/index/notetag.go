package index

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"strconv"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelastic "github.com/ryanreadbooks/whimer/misc/xelastic/analyzer"
	xelasticanalyzer "github.com/ryanreadbooks/whimer/misc/xelastic/analyzer"
	xelaserror "github.com/ryanreadbooks/whimer/misc/xelastic/errors"
	xelasformat "github.com/ryanreadbooks/whimer/misc/xelastic/format"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/tokenchar"
)

var _noteTagIns = NoteTag{}

// 笔记标签索引模型
type NoteTag struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Ctime int64  `json:"ctime"`
}

func (NoteTag) Index() string {
	return "note_tags"
}

func (NoteTag) AliasIndex() string {
	return "w_note_tags"
}

func (n NoteTag) Alias() map[string]types.Alias {
	return map[string]types.Alias{
		n.AliasIndex(): {},
	}
}

func NewCleanNormalizer() types.Normalizer {
	return &types.CustomNormalizer{
		Filter: []string{"lowercase", "asciifolding", "trim"},
	}
}

func (NoteTag) Settings(opt *NoteTagIndexerOption) *types.IndexSettings {
	return &types.IndexSettings{
		MaxNgramDiff: mg.Ptr(5),
		Analysis: &types.IndexSettingsAnalysis{
			// 自定义normalizer和tokenizer和analyzer
			Normalizer: map[string]types.Normalizer{
				"cust_clean_normalizer": NewCleanNormalizer(),
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

func (NoteTag) Mappings() *types.TypeMapping {
	return &types.TypeMapping{
		Dynamic: &dynamicmapping.True,
		Properties: map[string]types.Property{
			"id": types.NewKeywordProperty(),
			"name": &types.TextProperty{
				Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
				SearchAnalyzer: mg.Ptr(xelastic.IkSmart),
				Fields: map[string]types.Property{
					"keyword": &types.KeywordProperty{
						Normalizer: mg.Ptr("cust_clean_normalizer"),
					},
					"prefix": &types.TextProperty{
						Analyzer: mg.Ptr("cust_prefix_analyzer"),
					},
					"ngram": &types.TextProperty{
						Analyzer: mg.Ptr("cust_ngram_analyzer"),
					},
				},
			},
			"ctime": &types.DateProperty{
				Format: mg.Ptr(xelasformat.DateEpochSecond),
			},
		},
	}
}

type NoteTagIndexer struct {
	es *elasticsearch.TypedClient
}

func NewNoteTagIndexer(es *elasticsearch.TypedClient) *NoteTagIndexer {
	return &NoteTagIndexer{
		es: es,
	}
}

type NoteTagIndexerOption struct {
	NumberOfReplicas int
	NumbefOfShards   int
}

func fmtNoteTagDocId(n *NoteTag) string {
	return "note_tags:" + n.Id
}

// 初始化
func (n *NoteTagIndexer) Init(ctx context.Context, opt *NoteTagIndexerOption) error {
	exist, err := n.es.Indices.Exists(_noteTagIns.Index()).Do(ctx)
	if err != nil {
		return xelaserror.Convert(err)
	}
	if exist {
		return nil
	}

	_, err = n.es.Indices.
		Create(_noteTagIns.Index()).
		Mappings(_noteTagIns.Mappings()).
		Aliases(_noteTagIns.Alias()).
		Settings(_noteTagIns.Settings(opt)).
		Do(ctx)

	if err != nil {
		if xelaserror.IsResourceAlreadyExistsError(err) {
			return nil
		}

		return xelaserror.Convert(err)
	}

	return nil
}

func (n *NoteTagIndexer) Add(ctx context.Context, tag *NoteTag) error {
	_, err := n.es.Index(_noteTagIns.AliasIndex()).
		Id(fmtNoteTagDocId(tag)).
		Document(tag).
		Refresh(refresh.True).
		Do(ctx)

	if err != nil {
		return xelaserror.Convert(err)
	}

	return nil
}

func (n *NoteTagIndexer) BulkAdd(ctx context.Context, tags []*NoteTag) error {
	bulk, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:  n.es,
		Index:   _noteTagIns.AliasIndex(),
		Refresh: refresh.True.String(),
	})

	for _, tag := range tags {
		body, err := json.Marshal(tag)
		if err != nil {
			continue
		}

		err = bulk.Add(ctx, esutil.BulkIndexerItem{
			Index:      _noteTagIns.AliasIndex(),
			DocumentID: fmtNoteTagDocId(tag),
			Action:     "create",
			Body:       bytes.NewReader(body),
			OnFailure: func(ctx context.Context, bii esutil.BulkIndexerItem, biri esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					xlog.Msgf("note tag indexer item %s failed", bii.DocumentID).Err(err).Errorx(ctx)
				}
			},
		})

		if err != nil {
			xlog.Msgf("note tag indexer bulk add %s failed", tag.Id).Err(err).Errorx(ctx)
		}
	}

	err := bulk.Close(ctx)

	return xelaserror.Convert(err)
}

// 分页检索
// page starts from 1
// TODO 这个分页有问题
func (n *NoteTagIndexer) Search(ctx context.Context, text string, page, count int) ([]*NoteTag, int64, error) {
	from := (page - 1) * count
	from = min(from, math.MaxInt32)
	
	search := n.es.Search().Index(_noteTagIns.AliasIndex()).
		Query(&types.Query{
			Match: map[string]types.MatchQuery{
				"name.ngram": {
					Query:    text,
					Operator: &operator.And,
				},
			},
		}).
		From(from).
		Size(count)

	resp, err := search.Do(ctx)
	if err != nil {
		return nil, 0, xelaserror.Convert(err)
	}

	tags := make([]*NoteTag, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var hitTag NoteTag
		err = json.Unmarshal(hit.Source_, &hitTag)
		if err != nil {
			continue
		}

		tags = append(tags, &hitTag)
	}

	return tags, resp.Hits.Total.Value, nil
}
