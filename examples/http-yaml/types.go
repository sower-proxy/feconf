package main

// 共享的配置结构体
type Config struct {
	App      AppConfig      `json:"application"`    // Different tag name
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Features []string       `json:"feature_flags"`  // Different tag name
}

type AppConfig struct {
	Name        string `json:"app_name"`        // Different tag name
	Version     string `json:"version"`
	Environment string `json:"env"`             // Different tag name
}

type ServerConfig struct {
	Host  string `json:"bind_address"`         // Different tag name
	Port  int    `json:"port"`
	Debug bool   `json:"debug_mode"`           // Different tag name
}

type DatabaseConfig struct {
	Primary  DatabaseConnection `json:"primary"`
	Replicas []string           `json:"read_replicas"`  // Different tag name
}

type DatabaseConnection struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Name string `json:"database_name"`        // Different tag name
	URL  string `json:"connection_url"`       // Different tag name
}