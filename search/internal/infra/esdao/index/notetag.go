package index

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelaserror "github.com/ryanreadbooks/whimer/misc/xelastic/errors"
	xelasformat "github.com/ryanreadbooks/whimer/misc/xelastic/format"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
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
		n.AliasIndex(): {
		},
	}
}

func (NoteTag) Mappings() *types.TypeMapping {
	return &types.TypeMapping{
		Dynamic: &dynamicmapping.True,
		Properties: map[string]types.Property{
			"id":   types.NewKeywordProperty(),
			"name": types.NewTextProperty(),
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
		Settings(&types.IndexSettings{
			NumberOfReplicas: mg.Ptr(strconv.Itoa(opt.NumberOfReplicas)),
			NumberOfShards:   mg.Ptr(strconv.Itoa(opt.NumbefOfShards)),
		}).
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
