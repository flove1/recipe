package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HttpHost string `env:"HTTP_HOST" env-default:"localhost"`
	HttpPort int    `env:"HTTP_PORT" env-default:"8080"`

	Mongo DBConfig
	Redis RedisConfig
	Neo4j Neo4jConfig
}

type DBConfig struct {
	URL  string `env:"MONGO_URI" env-required:"true"`
	Name string `env:"MONGO_NAME" env-required:"true"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URI" env-required:"true"`
}

type Neo4jConfig struct {
	URL string `env:"NEO4J_URI" env-required:"true"`
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
