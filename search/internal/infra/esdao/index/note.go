package index

import (
	"context"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelasticanalyzer "github.com/ryanreadbooks/whimer/misc/xelastic/analyzer"
	xelaserror "github.com/ryanreadbooks/whimer/misc/xelastic/errors"
	xelasformat "github.com/ryanreadbooks/whimer/misc/xelastic/format"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
)

var _noteIns = Note{}

// 笔记索引模型
type Note struct {
	NoteId   string     `json:"note_id"`
	Title    string     `json:"title"`
	Desc     string     `json:"desc"`
	CreateAt int64      `json:"create_at"`
	UpdateAt int64      `json:"update_at"`
	Author   NoteAuthor `json:"author"`
	TagList  []*NoteTag `json:"tag_list"`
}

type NoteAuthor struct {
	Uid      int64  `json:"uid"`
	Nickname string `json:"nickname"`
}

func (Note) Index() string {
	return "notes"
}

func (Note) AliasIndex() string {
	return "w_notes"
}

func (n Note) Alias() map[string]types.Alias {
	return map[string]types.Alias{
		n.AliasIndex(): {},
	}
}

func (Note) Settings(opt *IndexerOption) *types.IndexSettings {
	return defaultSettings(opt)
}

func (Note) Mappings() *types.TypeMapping {
	return &types.TypeMapping{
		Dynamic: &dynamicmapping.True,
		Properties: map[string]types.Property{
			"note_id": types.NewKeywordProperty(),
			"title": &types.TextProperty{
				Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
				SearchAnalyzer: mg.Ptr(xelasticanalyzer.IkSmart),
				Fields:         defaultTextFields,
			},
			"desc": &types.TextProperty{
				Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
				SearchAnalyzer: mg.Ptr(xelasticanalyzer.IkSmart),
			},
			"create_at": &types.DateProperty{
				Format: mg.Ptr(xelasformat.DateEpochSecond),
			},
			"update_at": &types.DateProperty{
				Format: mg.Ptr(xelasformat.DateEpochSecond),
			},
			"author": &types.ObjectProperty{
				Properties: map[string]types.Property{
					"uid": types.NewLongNumberProperty(),
					"nickname": &types.TextProperty{
						Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
						SearchAnalyzer: mg.Ptr(xelasticanalyzer.IkSmart),
						Fields:         defaultTextFields,
					},
				},
			},
			"tag_list": &types.ObjectProperty{
				Properties: map[string]types.Property{
					"id": types.NewKeywordProperty(),
					"name": &types.TextProperty{
						Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
						SearchAnalyzer: mg.Ptr(xelasticanalyzer.IkSmart),
						Fields:         defaultTextFields,
					},
				},
			},
		},
	}
}

func (n *Note) GetId() string {
	return fmtNoteDocId(n)
}

func fmtNoteDocId(n *Note) string {
	return fmtNoteDocIdString(n.NoteId)
}

func fmtNoteDocIdString(noteId string) string {
	return "note:" + noteId
}

type NoteIndexer struct {
	es *elasticsearch.TypedClient
}

func NewNoteIndexer(es *elasticsearch.TypedClient) *NoteIndexer {
	return &NoteIndexer{
		es: es,
	}
}

func (n *NoteIndexer) Init(ctx context.Context, opt *IndexerOption) error {
	exist, err := n.es.Indices.Exists(_noteIns.Index()).Do(ctx)
	if err != nil {
		return xelaserror.Convert(err)
	}
	if exist {
		return nil
	}

	_, err = n.es.Indices.
		Create(_noteIns.Index()).
		Mappings(_noteIns.Mappings()).
		Aliases(_noteIns.Alias()).
		Settings(_noteIns.Settings(opt)).
		Do(ctx)
	if err != nil {
		if xelaserror.IsResourceAlreadyExistsError(err) {
			return nil
		}

		return xelaserror.Convert(err)
	}

	return nil
}

// 批量添加文档
func (n *NoteIndexer) BulkAdd(ctx context.Context, notes []*Note) error {
	return doBulkCreate(ctx, n.es, notes)
}

// 批量删除文档
func (n *NoteIndexer) BulkDelete(ctx context.Context, ids []string) error {
	noteDocIds := make([]string, 0, len(ids))
	for _, id := range ids {
		noteDocIds = append(noteDocIds, fmtNoteDocIdString(id))
	}
	return doBulkDelete(ctx, n.es, _noteIns.Index(), noteDocIds)
}
