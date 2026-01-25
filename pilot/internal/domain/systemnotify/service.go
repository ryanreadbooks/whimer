package systemnotify

import "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/repository"

type DomainService struct {
	systemNotifyAdapter repository.SystemNotifyAdapter
}

func NewDomainService(
	systemNotifyAdapter repository.SystemNotifyAdapter,
) *DomainService {
	return &DomainService{
		systemNotifyAdapter: systemNotifyAdapter,
	}
}

