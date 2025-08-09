package esdao

import (
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ryanreadbooks/whimer/search/internal/config"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index"
)

type EsDao struct {
	es *elasticsearch.TypedClient

	NoteTagIndexer *index.NoteTagIndexer
}

func MustNew(c *config.Config) *EsDao {
	addresses := strings.Split(c.ElasticSearch.Addr, ",")
	esc := elasticsearch.Config{
		Addresses: addresses,
		Username:  c.ElasticSearch.User,
		Password:  c.ElasticSearch.Password,
	}

	client, err := elasticsearch.NewTypedClient(esc)
	if err != nil {
		panic(err)
	}

	return &EsDao{
		es:             client,
		NoteTagIndexer: index.NewNoteTagIndexer(client),
	}
}

func (d *EsDao) Init() error {
	// 初始化索引
	return nil
}

func (d *EsDao) MustInit() {
	if err := d.Init(); err != nil {
		panic(err)
	}
}
