package polish

type ServerConfig interface {
	Construct() (Server, error)
	GetChunkSize() int
}

type ClientConfig interface {
	Construct() (Connection, error)
	GetChunkSize() int
}
