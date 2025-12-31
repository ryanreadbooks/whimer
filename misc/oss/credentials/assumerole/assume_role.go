package assumerole

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/signer"
	"github.com/zeromicro/go-zero/rest/httpc"
)

type STSAssumeRole struct {
	credentials.Expiry
	Client *http.Client
	Config Config
}

var _ Provider = &STSAssumeRole{}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string

	SessionToken string // Optional if the first request is made with temporary credentials.
	Policy       string // Optional to assign a policy to the assumed role

	Location        string // Optional commonly needed with AWS STS.
	DurationSeconds int
}

const defaultDurationSeconds = 900 // 15min

// closeResponse close non nil response with any response Body.
// convenient wrapper to drain any remaining data on response body.
//
// Subsequently this allows golang http RoundTripper
// to re-use the same connection for future requests.
func closeResponse(resp *http.Response) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if resp != nil && resp.Body != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		//  - http://stackoverflow.com/a/17961593/4465767
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func fetchAssumeRoleCredentials(ctx context.Context, client *http.Client, cfg Config) (*credentials.AssumeRoleResponse, error) {
	v := url.Values{}
	v.Set("Action", "AssumeRole")
	v.Set("Version", credentials.STSVersion)
	if cfg.DurationSeconds > defaultDurationSeconds {
		v.Set("DurationSeconds", strconv.Itoa(cfg.DurationSeconds))
	} else {
		v.Set("DurationSeconds", strconv.Itoa(defaultDurationSeconds))
	}
	if cfg.Policy != "" {
		v.Set("Policy", cfg.Policy)
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	u.Path = "/"

	postBody := strings.NewReader(v.Encode())
	hash := sha256.New()
	if _, err = io.Copy(hash, postBody); err != nil {
		return nil, err
	}
	postBody.Seek(0, 0)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), postBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(hash.Sum(nil)))
	if cfg.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", cfg.SessionToken)
	}
	req = signer.SignV4STS(*req, cfg.AccessKey, cfg.SecretKey, cfg.Location)
	service := httpc.NewServiceWithClient("oss-credentials-service", client)
	resp, err := service.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		var errResp credentials.ErrorResponse
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		_, err = xmlDecodeAndBody(bytes.NewReader(buf), &errResp)
		if err != nil {
			var s3Err credentials.Error
			if _, err = xmlDecodeAndBody(bytes.NewReader(buf), &s3Err); err != nil {
				return nil, err
			}
			errResp.RequestID = s3Err.RequestID
			errResp.STSError.Code = s3Err.Code
			errResp.STSError.Message = s3Err.Message
		}
		return nil, errResp
	}

	a := credentials.AssumeRoleResponse{}
	if _, err = xmlDecodeAndBody(resp.Body, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// Retrieve retrieves credentials from the MinIO service.
// Error will be returned if the request fails.
func (s *STSAssumeRole) Retrieve(ctx context.Context) (credentials.Value, error) {
	a, err := fetchAssumeRoleCredentials(ctx, s.Client, s.Config)
	if err != nil {
		return credentials.Value{}, err
	}

	// Expiry window is set to 10secs.
	s.SetExpiration(a.Result.Credentials.Expiration, credentials.DefaultExpiryWindow)

	return credentials.Value{
		AccessKeyID:     a.Result.Credentials.AccessKey,
		SecretAccessKey: a.Result.Credentials.SecretKey,
		SessionToken:    a.Result.Credentials.SessionToken,
		Expiration:      a.Result.Credentials.Expiration,
		SignerType:      credentials.SignatureV4,
	}, nil
}

type stsAssumeRoleExtra struct {
	httpCli *http.Client
}

type STSAssumeRoleExtra func(*stsAssumeRoleExtra)

func WithExtraClient(c *http.Client) STSAssumeRoleExtra {
	return func(
		s *stsAssumeRoleExtra) {
		s.httpCli = c
	}
}

// NewSTSAssumeRole returns a pointer to a new
// Credentials object wrapping the STSAssumeRole.
func NewSTSAssumeRole(cfg Config, extras ...STSAssumeRoleExtra) (*Credentials, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("STS endpoint cannot be empty")
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, errors.New("AssumeRole credentials access/secretkey is mandatory")
	}
	ar := &STSAssumeRole{
		Config: cfg,
	}

	defaultExtra := &stsAssumeRoleExtra{
		httpCli: &http.Client{
			Transport: http.DefaultTransport,
		},
	}

	for _, apply := range extras {
		apply(defaultExtra)
	}
	ar.Client = defaultExtra.httpCli

	return New(ar), nil
}
