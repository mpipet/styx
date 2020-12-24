package nodeman

var (
	DefaultConfig = Config{
		NodeName:         "node",
		RaftDirectory:    "./raft",
		AdvertiseAddress: "127.0.0.1:8000",
	}
)

type Config struct {
	NodeName         string
	RaftDirectory    string
	AdvertiseAddress string
}
