package statsd

var DefaultConfig = Config{
	Address: "localhost:8125",
	Prefix:  "styx",
}

type Config struct {
	Protocol string
	Address  string
	Prefix   string
}
