package index

import (
	"context"
	"encoding/json"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelaserror "github.com/ryanreadbooks/whimer/misc/xelastic/errors"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type AliasHolder interface {
	AliasIndex() string
}

type BulkCreatedInstance interface {
	AliasHolder
	GetId() string
}

func doBulkCreate[T BulkCreatedInstance](ctx context.Context, es *elasticsearch.TypedClient, ins []T) error {
	if len(ins) == 0 {
		return nil
	}

	bulk := es.Bulk().Index(ins[0].AliasIndex())

	for _, i := range ins {
		body, err := json.Marshal(i)
		if err != nil {
			continue
		}

		cop := types.NewCreateOperation()
		cop.Id_ = mg.Ptr(i.GetId())
		err = bulk.CreateOp(*cop, body)
		if err != nil {
			return xelaserror.Convert(err)
		}
	}

	resp, err := bulk.Do(ctx)
	if err != nil {
		return xelaserror.Convert(err)
	}

	if err := handleBulkResponse(ctx, resp); err != nil {
		return err
	}

	return nil
}

func doBulkDelete(ctx context.Context, es *elasticsearch.TypedClient, index string, docIds []string) error {
	if len(docIds) == 0 {
		return nil
	}

	bulk := es.Bulk().Index(index)
	for _, id := range docIds {
		ope := types.NewDeleteOperation()
		ope.Id_ = mg.Ptr(id)
		err := bulk.DeleteOp(*ope)
		if err != nil {
			return xelaserror.Convert(err)
		}
	}

	resp, err := bulk.Do(ctx)
	if err != nil {
		return xelaserror.Convert(err)
	}

	if err := handleBulkResponse(ctx, resp); err != nil {
		return err
	}

	return nil
}
