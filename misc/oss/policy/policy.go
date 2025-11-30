package policy

import (
	"encoding/json"
	"fmt"

	v7policy "github.com/minio/minio-go/v7/pkg/policy"
	"github.com/minio/minio-go/v7/pkg/set"
)

const (
	Version = "2012-10-17"
)

const (
	EffectAllow = "Allow"
	EffectDeny  = "Deny"
)

// 简化policy的使用
type Policy struct {
	bp *v7policy.BucketAccessPolicy
}

func New() *Policy {
	return &Policy{
		bp: &v7policy.BucketAccessPolicy{
			Version:    Version,
			Statements: []v7policy.Statement{},
		},
	}
}

func (p *Policy) AppendStatement(stmt v7policy.Statement) {
	p.bp.Statements = append(p.bp.Statements, stmt)
}

func NewAllowStatement() v7policy.Statement {
	return v7policy.Statement{
		Effect:     EffectAllow,
		Actions:    set.NewStringSet(),
		Resources:  set.NewStringSet(),
		Principal:  v7policy.User{AWS: set.CreateStringSet("*")},
		Conditions: make(v7policy.ConditionMap),
	}
}

func (p *Policy) String() string {
	c, _ := json.Marshal(p.bp)
	return string(c)
}

// bucket + prefix
func GetSimpleResource(key string) string {
	return fmt.Sprintf("arn:aws:s3:::%s", key)
}
