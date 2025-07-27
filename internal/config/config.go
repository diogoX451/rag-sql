package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB    DatabaseConfig
	Neo4j Neo4jConfig
}

type Neo4jConfig struct {
	URI      string
	User     string
	Password string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() (*Config, error) {
	db := DatabaseConfig{
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "5432"),
		User:     getenv("DB_USER", "postgres"),
		Password: getenv("DB_PASSWORD", "postgres"),
		Name:     getenv("DB_NAME", "postgres"),
		SSLMode:  getenv("DB_SSLMODE", "disable"),
	}

	neo4j := Neo4jConfig{
		URI:      getenv("NEO4J_URI", "bolt://localhost:7687"),
		User:     getenv("NEO4J_USER", "neo4j"),
		Password: getenv("NEO4J_PASSWORD", "your_password"),
	}

	return &Config{DB: db, Neo4j: neo4j}, nil
}

func (d DatabaseConfig) ConnString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func getenv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
