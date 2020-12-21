package nodeman

var (
	DefaultConfig = Config{
		NodeName:       "node",
		StateDirectory: "./state",
		RaftAddress:    "127.0.0.1:11000",
	}
)

type Config struct {
	NodeName       string
	StateDirectory string
	RaftAddress    string
}
