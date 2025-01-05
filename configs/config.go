package configs

type Config struct {
	Server struct {
		Address string
		Port    int
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
}

func Load() (*Config, error) {
	// Load configuration from environment variables or config file
	// Implementation details here
	return &Config{}, nil
}
