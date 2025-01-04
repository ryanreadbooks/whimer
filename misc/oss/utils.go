package oss

import "fmt"

func GetPublicVisitUrl(bucket, objectName, endpoint string) string {
	return fmt.Sprintf("%s/%s/%s", endpoint, bucket, objectName)
}

func GetPublicVisitUrl2(bObjName, endpoint string) string {
	return fmt.Sprintf("%s/%s", endpoint, bObjName)
}
