package condition

// conditions
const (
	StringEquals = "StringEquals"
)

// condition keys
const (
	S3SignatureVersion     = "s3:signatureversion"
	S3RequestObjectTagKeys = "s3:RequestObjectTagKeys"
	S3RequestObjectTag     = "s3:RequestObjectTag"
)

// some condition values
const (
	// https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/bucket-policy-s3-sigv4-conditions.html
	SignatureV4 = "AWS4-HMAC-SHA256"
)
