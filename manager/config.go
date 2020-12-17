package manager

var (
	DefaultConfig = Config{
		DataDirectory:   "./data",
		WriteBufferSize: 1 << 20, // 1MB
	}
)

type Config struct {
	DataDirectory   string
	WriteBufferSize int
}
