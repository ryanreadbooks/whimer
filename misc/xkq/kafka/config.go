package kafka

type Config struct {
	Brokers  string `json:"brokers"`
	Username string `json:"username"`
	Password string `json:"password"`
}
