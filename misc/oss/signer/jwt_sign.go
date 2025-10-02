package signer

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

type JwtSignConfig struct {
	JwtIssuer   string        `json:"jwt_issuer"`
	JwtSubject  string        `json:"jwt_subject"`
	JwtDuration time.Duration `json:"jwt_duration"`
	Ak          string        `json:"ak"`
	Sk          string        `json:"sk"`
}

type JwtUploadAuthSigner struct {
	c *JwtSignConfig
}

type STSUploadAuthClaim struct {
	jwtv5.RegisteredClaims

	AccessKey string   `json:"access_key"`
	FileIds   []string `json:"file_ids"`
	Resource  string   `json:"resource"`
}

func newSTSUploadAuthClaim(c *JwtSignConfig, fileIds []string, resource string, now, expireAt time.Time) (
	*STSUploadAuthClaim, error) {
	akb := make([]byte, 16)
	_, err := rand.Read(akb)
	if err != nil {
		return nil, err
	}

	hash := hmac.New(sha256.New, []byte(c.Sk))
	hash.Write(akb)
	akb = hash.Sum(nil)
	ak := hex.EncodeToString(akb)

	return &STSUploadAuthClaim{
		AccessKey: ak,
		FileIds:   fileIds,
		Resource:  resource,

		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    c.JwtIssuer,
			Subject:   c.JwtSubject,
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(expireAt),
		},
	}, nil
}

func NewJwtUploadAuthSigner(c *JwtSignConfig) *JwtUploadAuthSigner {
	return &JwtUploadAuthSigner{c: c}
}

type JwtSignedUploadAuth struct {
	CurrentTime int64
	ExpireTime  int64
	Token       string
}

func (s *JwtUploadAuthSigner) GetUploadAuth(fileId, resource string) (JwtSignedUploadAuth, error) {
	return s.BatchGetUploadAuth([]string{fileId}, resource)
}

func (s *JwtUploadAuthSigner) BatchGetUploadAuth(fileIds []string, resource string) (JwtSignedUploadAuth, error) {
	var res JwtSignedUploadAuth

	now := time.Now()
	expireAt := now.Add(s.c.JwtDuration)
	claim, err := newSTSUploadAuthClaim(s.c, fileIds, resource, now, expireAt)
	if err != nil {
		return res, err
	}

	jwtToken := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claim)
	ss, err := jwtToken.SignedString(xstring.AsBytes(s.c.Sk))

	if err != nil {
		return res, err
	}

	res.CurrentTime = now.Unix()
	res.ExpireTime = expireAt.Unix()
	res.Token = ss
	return res, nil
}
