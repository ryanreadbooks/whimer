package note

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelasticanalyzer "github.com/ryanreadbooks/whimer/misc/xelastic/analyzer"
	xelaserror "github.com/ryanreadbooks/whimer/misc/xelastic/errors"
	xelasformat "github.com/ryanreadbooks/whimer/misc/xelastic/format"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index/common"
	"github.com/ryanreadbooks/whimer/search/pkg"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/fieldtype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/vmihailenco/msgpack/v5"
)

var _noteIns = Note{}

// 笔记索引模型
type Note struct {
	NoteId        string             `json:"note_id"`
	Title         string             `json:"title"`
	Desc          string             `json:"desc"`
	CreateAt      int64              `json:"create_at"`
	UpdateAt      int64              `json:"update_at"`
	Author        NoteAuthor         `json:"author"`
	TagList       []*NoteTag         `json:"tag_list"`
	AssetType     pkg.NoteAssetType  `json:"asset_type"`
	Visibility    pkg.NoteVisibility `json:"visibility"`
	LikesCount    int64              `json:"likes_count"`
	CommentsCount int64              `json:"comments_count"`
}

// 搜索仅返回notId
type NoteSearchResult struct {
	NoteId string `json:"note_id"`
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

func (Note) Settings(opt *common.IndexerOption) *types.IndexSettings {
	return common.DefaultSettings(opt)
}

func (Note) Mappings() *types.TypeMapping {
	return &types.TypeMapping{
		Dynamic: &dynamicmapping.True,
		Properties: map[string]types.Property{
			"note_id": types.NewKeywordProperty(),
			"title": &types.TextProperty{
				Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
				SearchAnalyzer: mg.Ptr(xelasticanalyzer.IkSmart),
				Fields:         common.DefaultTextFields,
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
						Fields:         common.DefaultTextFields,
					},
				},
			},
			"tag_list": &types.ObjectProperty{
				Properties: map[string]types.Property{
					"id": types.NewKeywordProperty(),
					"name": &types.TextProperty{
						Analyzer:       mg.Ptr(xelasticanalyzer.IkMaxWord),
						SearchAnalyzer: mg.Ptr(xelasticanalyzer.IkSmart),
						Fields:         common.DefaultTextFields,
					},
				},
			},
			"asset_type":     types.NewKeywordProperty(),
			"visibility":     types.NewKeywordProperty(),
			"likes_count":    types.NewLongNumberProperty(),
			"comments_count": types.NewLongNumberProperty(),
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

func (n *NoteIndexer) Init(ctx context.Context, opt *common.IndexerOption) error {
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

func (n *NoteIndexer) BulkRequest(ctx context.Context, reqs []NoteAction) error {
	if len(reqs) == 0 {
		return nil
	}

	bulk := n.es.Bulk().Index(_noteIns.AliasIndex())

	// add bulk operations item
	for _, req := range reqs {
		switch req.Type() {
		case ActionCreateNote:
			body, err := req.GetDoc()
			if err != nil {
				continue
			}

			// 存在则更新 不存在则创建
			iop := types.NewIndexOperation()
			iop.Id_ = mg.Ptr(req.GetDocId())
			err = bulk.IndexOp(*iop, body)
			if err != nil {
				continue
			}

		case ActionDeleteNote:
			dop := types.NewDeleteOperation()
			dop.Id_ = mg.Ptr(req.GetDocId())
			err := bulk.DeleteOp(*dop)
			if err != nil {
				continue
			}
		}
	}

	resp, err := bulk.Do(ctx)
	if err != nil {
		return xelaserror.Convert(err)
	}

	if err := common.HandleBulkResponse(ctx, resp); err != nil {
		return err
	}

	return nil
}

// 批量添加文档
func (n *NoteIndexer) BulkAdd(ctx context.Context, notes []*Note) error {
	return common.DoBulkCreate(ctx, n.es, notes)
}

// 批量删除文档
func (n *NoteIndexer) BulkDelete(ctx context.Context, ids []string) error {
	noteDocIds := make([]string, 0, len(ids))
	for _, id := range ids {
		noteDocIds = append(noteDocIds, fmtNoteDocIdString(id))
	}
	return common.DoBulkDelete(ctx, n.es, _noteIns.AliasIndex(), noteDocIds)
}

var (
	// search querys
	noteVisibilityFilter = types.Query{
		Term: map[string]types.TermQuery{"visibility": {Value: pkg.VisibilityPublic}},
	}

	// we only return note_id in note searching
	noteSearchResultIncludes = types.SourceFilter{
		Includes: []string{"note_id"},
	}
)

// note search sort releated vars
var (
	noteDefaultSortByNoteId = types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"note_id": {
				Order: &sortorder.Desc,
			},
		},
	}

	noteDefaultSortByScore = types.SortOptions{
		Score_: &types.ScoreSort{Order: &sortorder.Desc},
	}

	noteDefaultSortByUpdateAt = types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"update_at": {
				Order:        &sortorder.Desc,
				UnmappedType: &fieldtype.Date,
			},
		},
	}

	noteDefaultSortByCreateAt = types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"create_at": {
				Order:        &sortorder.Desc,
				UnmappedType: &fieldtype.Date,
			},
		},
	}

	noteDefaultSortByLikesCount = types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"likes_count": {
				Order:        &sortorder.Desc,
				UnmappedType: &fieldtype.Long,
			},
		},
	}

	noteDefaultSortByCommentsCount = types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"comments_count": {
				Order:        &sortorder.Desc,
				UnmappedType: &fieldtype.Date,
			},
		},
	}

	// default note sort (sort by relevance)
	noteDefaultSort = []types.SortCombinations{
		noteDefaultSortByScore,
		noteDefaultSortByUpdateAt,
		noteDefaultSortByNoteId,
	}
)

const (
	noteTitleBoost       float32 = 5.0
	noteTitleNgramBoost  float32 = noteTitleBoost * 0.85
	noteTitlePrefixBoost float32 = noteTitleBoost * 0.85
	noteTagListBoost     float32 = 3.5
	noteAuthorBoost      float32 = 0.9
)

type NoteIndexSearchResult struct {
	NoteIds   []string
	Total     int64
	NextToken string
	HasNext   bool
}

type searchNoteOption struct {
	sortBy         string
	filterNoteType string
	filterPubTime  string
}

// 排序规则
func (o *searchNoteOption) sortRule() []types.SortCombinations {
	switch o.sortBy {
	case pkg.NoteFilterSortByRelevance:
		return noteDefaultSort
	case pkg.NoteFilterSortByNewest:
		return []types.SortCombinations{
			noteDefaultSortByCreateAt,
			noteDefaultSortByNoteId,
		}
	case pkg.NoteFilterSortByLikes:
		return []types.SortCombinations{
			noteDefaultSortByLikesCount,
			noteDefaultSortByUpdateAt,
			noteDefaultSortByNoteId,
		}
	case pkg.NoteFilterSortByComments:
		return []types.SortCombinations{
			noteDefaultSortByCommentsCount,
			noteDefaultSortByUpdateAt,
			noteDefaultSortByNoteId,
		}
	}

	return noteDefaultSort
}

func (searchNoteOption) assetTypeFilter(assetType string) types.Query {
	return types.Query{
		Terms: &types.TermsQuery{
			TermsQuery: map[string]types.TermsQueryField{
				"asset_type": []types.FieldValue{assetType},
			},
		},
	}
}

func (o *searchNoteOption) filterRule() []types.Query {
	filters := []types.Query{}
	// note type filters
	switch o.filterNoteType {
	case pkg.NoteFilterNoteTypeImage:
		filters = append(filters, o.assetTypeFilter(pkg.NoteFilterNoteTypeImage))
	}

	// note pub time filter
	startAt := int64(0)
	endAt := time.Now()
	switch o.filterPubTime {
	case pkg.NoteFilterPubTimeInOneDay: // 1 day
		startAt = endAt.Add(-time.Hour * 24).Unix()
	case pkg.NoteFilterPubTimeInOneWeek: // 7 days
		startAt = endAt.Add(-time.Hour * 24 * 7).Unix()
	case pkg.NoteFilterPubTimeInHalfYear: // 180 days
		startAt = endAt.Add(-time.Hour * 24 * 30 * 6).Unix()
	}
	if startAt > 0 {
		filters = append(filters, types.Query{
			Range: map[string]types.RangeQuery{
				"create_at": types.NumberRangeQuery{
					Gte: mg.Ptr(types.Float64(startAt)),
					Lte: mg.Ptr(types.Float64(endAt.Unix())),
				},
			},
		})
	}

	return filters
}

type SearchNoteOption func(o *searchNoteOption)

func newDefaultSearchNoteOption() *searchNoteOption {
	return &searchNoteOption{
		sortBy:         pkg.NoteFilterSortByRelevance,
		filterNoteType: pkg.NoteFilterNoteTypeAll,
		filterPubTime:  pkg.NoteFilterPubTimeAll,
	}
}

func WithSearchNoteSortBy(sortBy string) SearchNoteOption {
	return func(o *searchNoteOption) {
		o.sortBy = sortBy
	}
}

func WithSearchNoteSortByNewest() SearchNoteOption {
	return WithSearchNoteSortBy(pkg.NoteFilterSortByNewest)
}

func WithSearchNoteSortByLikes() SearchNoteOption {
	return WithSearchNoteSortBy(pkg.NoteFilterSortByLikes)
}

func WithSearchNoteSortByComments() SearchNoteOption {
	return WithSearchNoteSortBy(pkg.NoteFilterSortByComments)
}

func WithSearchNoteFilterNoteType(s string) SearchNoteOption {
	return func(o *searchNoteOption) {
		o.filterNoteType = s
	}
}

func WithSearchNoteImageOnly() SearchNoteOption {
	return WithSearchNoteFilterNoteType(pkg.NoteFilterNoteTypeImage)
}

func WithSearchNotePubTime(s string) SearchNoteOption {
	return func(o *searchNoteOption) {
		o.filterPubTime = s
	}
}

func WithSearchNotePubInOneDay() SearchNoteOption {
	return WithSearchNotePubTime(pkg.NoteFilterPubTimeInOneDay)
}

func WithSearchNotePubInOneWeek() SearchNoteOption {
	return WithSearchNotePubTime(pkg.NoteFilterPubTimeInOneWeek)
}

func WithSearchNotePubInHalfYear() SearchNoteOption {
	return WithSearchNotePubTime(pkg.NoteFilterPubTimeInHalfYear)
}

func (n *NoteIndexer) Search(ctx context.Context, keyword, pageToken string, count int32, opts ...SearchNoteOption) (*NoteIndexSearchResult, error) {
	opt := newDefaultSearchNoteOption()
	for _, o := range opts {
		o(opt)
	}

	// title related query
	titleQuery := []types.Query{
		{
			Match: map[string]types.MatchQuery{
				"title": {
					Query: keyword,
					Boost: mg.Ptr(noteTitleBoost),
				},
			},
		},
		{
			Match: map[string]types.MatchQuery{
				"title.ngram": {
					Query:    keyword,
					Boost:    mg.Ptr(noteTitleNgramBoost),
					Operator: &operator.And,
				},
			},
		},
		{
			Match: map[string]types.MatchQuery{
				"title.prefix": {
					Query: keyword,
					Boost: mg.Ptr(noteTitlePrefixBoost),
				},
			},
		}}

	// tag list related query
	tagListQuery := []types.Query{
		{
			Match: map[string]types.MatchQuery{
				"tag_list.name": {
					Query: keyword,
					Boost: mg.Ptr(noteTagListBoost),
				},
			},
		},
		{
			Match: map[string]types.MatchQuery{
				"tag_list.name.ngram": {
					Query:    keyword,
					Boost:    mg.Ptr(noteTagListBoost),
					Analyzer: mg.Ptr(common.CustomNgramAnalyzer),
					Operator: &operator.And,
				},
			},
		},
	}

	// author related query
	authorQuery := []types.Query{{
		Match: map[string]types.MatchQuery{
			"author.nickname": {
				Query: keyword,
				Boost: mg.Ptr(noteAuthorBoost),
			},
		},
	}}

	shouldQuery := []types.Query{}
	shouldQuery = append(shouldQuery, titleQuery...)
	shouldQuery = append(shouldQuery, tagListQuery...)
	shouldQuery = append(shouldQuery, authorQuery...)

	filters := []types.Query{noteVisibilityFilter}
	filters = append(filters, opt.filterRule()...)

	boolQuery := types.BoolQuery{
		Filter:             filters,
		Should:             shouldQuery,
		MinimumShouldMatch: 1, // should中的条件至少要满足一个
	}

	query := n.es.Search().
		Index(_noteIns.AliasIndex()).
		Query(&types.Query{
			Bool: &boolQuery,
		}).
		Sort(opt.sortRule()...).
		Source_(noteSearchResultIncludes).
		TrackTotalHits(true).
		Size(int(count))

	// try parse next token
	if len(pageToken) > 0 {
		sa := _noteIns.calculateSearchAfter(pageToken)
		if len(sa) > 0 {
			query.SearchAfter(sa...)
		}
	}

	resp, err := query.Do(ctx)
	if err != nil {
		return nil, xelaserror.Convert(err)
	}

	hitsLen := len(resp.Hits.Hits)
	if hitsLen == 0 {
		return &NoteIndexSearchResult{}, nil
	}

	noteIds := make([]string, 0, hitsLen)
	for _, hit := range resp.Hits.Hits {
		var nid NoteSearchResult
		err := json.Unmarshal(hit.Source_, &nid)
		if err == nil {
			noteIds = append(noteIds, nid.NoteId)
		}
	}

	// calculate next page token here
	lsr := resp.Hits.Hits[hitsLen-1].Sort
	nextToken := _noteIns.calculateNextToken(lsr)

	return &NoteIndexSearchResult{
		NoteIds:   noteIds,
		NextToken: nextToken,
		Total:     resp.Hits.Total.Value,
		HasNext:   len(noteIds) == int(count),
	}, nil
}

func (Note) calculateNextToken(lsr []types.FieldValue) string {
	if len(lsr) == 0 {
		return ""
	}
	data, err := msgpack.Marshal(lsr)
	if err != nil {
		return ""
	}

	return base64.RawStdEncoding.EncodeToString(data)
}

func (Note) calculateSearchAfter(s string) []types.FieldValue {
	data, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return nil
	}

	res := make([]types.FieldValue, 0)
	err = msgpack.Unmarshal(data, &res)
	if err != nil {
		return nil
	}

	return res
}
