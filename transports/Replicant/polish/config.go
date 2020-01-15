package polish

type ServerConfig interface {
	Construct() (Server, error)
}

type ClientConfig interface {
	Construct() (Connection, error)
}
