package note

import (
	"context"
	"encoding/json"
	"math"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelasticanalyzer "github.com/ryanreadbooks/whimer/misc/xelastic/analyzer"
	xelaserror "github.com/ryanreadbooks/whimer/misc/xelastic/errors"
	xelasformat "github.com/ryanreadbooks/whimer/misc/xelastic/format"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index/common"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
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

func (NoteTag) Settings(opt *common.IndexerOption) *types.IndexSettings {
	return common.DefaultSettings(opt)
}

func (NoteTag) Mappings() *types.TypeMapping {
	return &types.TypeMapping{
		Dynamic: &dynamicmapping.True,
		Properties: map[string]types.Property{
			"id": types.NewKeywordProperty(),
			"name": &types.TextProperty{
				Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
				SearchAnalyzer: mg.Ptr(xelasticanalyzer.IkSmart),
				Fields:         common.DefaultTextFields,
			},
			"ctime": &types.DateProperty{
				Format: mg.Ptr(xelasformat.DateEpochSecond),
			},
		},
	}
}

func (n *NoteTag) GetId() string {
	return fmtNoteTagDocId(n)
}

type NoteTagIndexer struct {
	es *elasticsearch.TypedClient
}

func NewNoteTagIndexer(es *elasticsearch.TypedClient) *NoteTagIndexer {
	return &NoteTagIndexer{
		es: es,
	}
}

func fmtNoteTagDocId(n *NoteTag) string {
	return "note_tags:" + n.Id
}

// 初始化
func (n *NoteTagIndexer) Init(ctx context.Context, opt *common.IndexerOption) error {
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
	return common.DoBulkCreate(ctx, n.es, tags)
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
