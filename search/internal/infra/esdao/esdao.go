package esdao

import (
	"context"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ryanreadbooks/whimer/search/internal/config"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index/common"
	indexnote "github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index/note"

	"go.opentelemetry.io/otel"
)

type EsDao struct {
	es *elasticsearch.TypedClient

	NoteTagIndexer *indexnote.NoteTagIndexer
	NoteIndexer    *indexnote.NoteIndexer
}

func MustNew(c *config.Config) *EsDao {
	addresses := strings.Split(c.ElasticSearch.Addr, ",")

	esc := elasticsearch.Config{
		Addresses:       addresses,
		Username:        c.ElasticSearch.User,
		Password:        c.ElasticSearch.Password,
		Instrumentation: elasticsearch.NewOpenTelemetryInstrumentation(otel.GetTracerProvider(), false),
	}

	client, err := elasticsearch.NewTypedClient(esc)
	if err != nil {
		panic(err)
	}

	return &EsDao{
		es:             client,
		NoteTagIndexer: indexnote.NewNoteTagIndexer(client),
		NoteIndexer:    indexnote.NewNoteIndexer(client),
	}
}

func (d *EsDao) Init(c *config.Config) error {
	// 初始化索引
	ctx := context.Background()
	err := d.NoteTagIndexer.Init(ctx, &common.IndexerOption{
		NumberOfReplicas: c.Indices.NoteTag.NumReplicas,
		NumbefOfShards:   c.Indices.NoteTag.NumShards,
	})
	if err != nil {
		return err
	}
	err = d.NoteIndexer.Init(ctx, &common.IndexerOption{
		NumberOfReplicas: c.Indices.Note.NumReplicas,
		NumbefOfShards:   c.Indices.Note.NumShards,
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *EsDao) MustInit(c *config.Config) {
	if err := d.Init(c); err != nil {
		panic(err)
	}
}
