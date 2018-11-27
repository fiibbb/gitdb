package config

type AppConfig struct {
	GRPCAddr string
	HTTPAddr string
	RepoPath string
}

func NewAppConfig() (*AppConfig, error) {
	return &AppConfig{
		GRPCAddr: "localhost:8080",
		HTTPAddr: "localhost:8081",
		RepoPath: "/Users/lingy/Uber/tmp/repo2/repo.git",
	}, nil
}
