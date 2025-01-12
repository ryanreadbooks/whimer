package model

import (
	"github.com/minio/minio-go/v7/pkg/notification"
)

type MinioEvent struct {
	EventName string                `json:"EventName"`
	Key       string                `json:"Key"`
	Records   []*notification.Event `json:"Records"`
}
