package index

import (
	"context"
	"encoding/json"

	mg "github.com/ryanreadbooks/whimer/misc/generics"
	xelaserror "github.com/ryanreadbooks/whimer/misc/xelastic/errors"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type BulkCreatedInstance interface {
	GetId() string
	AliasIndex() string
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
		err = bulk.CreateOp(*types.NewCreateOperation(), body)
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
