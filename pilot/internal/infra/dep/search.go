package dep

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

var (
	searcher   searchv1.SearchServiceClient
	documenter searchv1.DocumentServiceClient
)

func InitSearch(c *config.Config) {
	searcher = xgrpc.NewRecoverableClient(c.Backend.Search,
		searchv1.NewSearchServiceClient,
		func(cc searchv1.SearchServiceClient) { searcher = cc })

	documenter = xgrpc.NewRecoverableClient(c.Backend.Search,
		searchv1.NewDocumentServiceClient,
		func(cc searchv1.DocumentServiceClient) { documenter = cc })

}

func SearchServer() searchv1.SearchServiceClient {
	return searcher
}

func DocumentServer() searchv1.DocumentServiceClient {
	return documenter
}
