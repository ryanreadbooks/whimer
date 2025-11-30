package credentials

import "github.com/ryanreadbooks/whimer/misc/oss/credentials/assumerole"

type STSCredentials = assumerole.Credentials

type Config struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	Policy          string
	DurationSeconds int
}

func NewSTSCredentials(cfg Config) (*STSCredentials, error) {
	c, err := assumerole.NewSTSAssumeRole(assumerole.Config{
		Endpoint:        cfg.Endpoint,
		AccessKey:       cfg.AccessKey,
		SecretKey:       cfg.SecretKey,
		Policy:          cfg.Policy,
		DurationSeconds: cfg.DurationSeconds,
	})
	return c, err
}
