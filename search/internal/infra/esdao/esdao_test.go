package esdao

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xelastic/format"
	"github.com/ryanreadbooks/whimer/search/internal/config"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index/common"
	noteindex "github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index/note"

	mg "github.com/ryanreadbooks/whimer/misc/generics"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

var testEsDao *EsDao

func TestMain(m *testing.M) {
	testEsDao = MustNew(&config.Config{
		ElasticSearch: config.ElasticSearch{
			Addr:     os.Getenv("ENV_ES_ADDR"),
			User:     os.Getenv("ENV_ES_USER"),
			Password: os.Getenv("ENV_ES_PASSWORD"),
		},
	})

	m.Run()
}

func TestNew(t *testing.T) {
	resp, err := testEsDao.es.Info().Do(context.TODO())
	t.Log(err)
	t.Log(resp)
}

func TestMappings(t *testing.T) {
	p := types.NewDateProperty()
	p.Format = mg.Ptr(format.DateEpochSecond)
	c, _ := json.Marshal(p)
	t.Log(string(c))
}

func TestNoteTagIndex(t *testing.T) {
	ctx := context.TODO()
	indx := noteindex.NewNoteTagIndexer(testEsDao.es)
	err := indx.Init(ctx, &common.IndexerOption{
		NumberOfReplicas: 0,
		NumbefOfShards:   1,
	})
	t.Log(err)

	err = indx.Add(ctx, &noteindex.NoteTag{
		Id:    "test_abc",
		Name:  "test_name",
		Ctime: time.Now().Unix(),
	})
	t.Log(err)

	err = indx.BulkAdd(ctx, []*noteindex.NoteTag{
		{
			Id:    "test_abc",
			Name:  "test_name2",
			Ctime: time.Now().Unix(),
		},
		{
			Id:    "testing",
			Name:  "testing",
			Ctime: time.Now().Unix(),
		}})
	t.Log(err)
}
