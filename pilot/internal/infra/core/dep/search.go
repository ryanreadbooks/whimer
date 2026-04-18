package dep

import (
	searchv1 "github.com/ryanreadbooks/whimer/idl/gen/go/search/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

var (
	searcher   searchv1.SearchServiceClient
	documenter searchv1.DocumentServiceClient
)

func InitSearch(c *config.Config) {
	conn := xgrpc.NewRecoverableClientConn(c.Backend.Search)
	searcher = searchv1.NewSearchServiceClient(conn)
	documenter = searchv1.NewDocumentServiceClient(conn)

}

func SearchServer() searchv1.SearchServiceClient {
	return searcher
}

func DocumentServer() searchv1.DocumentServiceClient {
	return documenter
}
