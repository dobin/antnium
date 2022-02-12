package server

type Config struct {
	AdminApiKey string
}

func MakeConfig() Config {
	c := Config{
		"Secret-AdminApi-Key",
	}
	return c
}
