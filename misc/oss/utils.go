package oss

import "fmt"

func GetPublicVisitUrl(bucket, objectName, endpoint string) string {
	return fmt.Sprintf("%s/%s/%s", endpoint, bucket, objectName)
}
