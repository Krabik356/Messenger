package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort   string
	TokensSecret string
	Redis        struct {
		RedisAddress string
		RedisPort    string
	}
	Postgres struct {
		PostgresUser         string
		PostgresPassword     string
		PostgresAddress      string
		PostgresPort         string
		PostgresDatabaseName string
	}
	Producer struct {
		ProcuerAddress string
		ProcuerPort    string
	}
}

func NewConfigWithDataFromEnv() *Config {
	return &Config{
		ServerPort:   os.Getenv("SERVERPORT"),
		TokensSecret: os.Getenv("TOKENSSECRET"),
		Redis: struct {
			RedisAddress string
			RedisPort    string
		}{
			RedisAddress: os.Getenv("REDISADDRESS"),
			RedisPort:    os.Getenv("REDISPORT"),
		},
		Postgres: struct {
			PostgresUser         string
			PostgresPassword     string
			PostgresAddress      string
			PostgresPort         string
			PostgresDatabaseName string
		}{
			PostgresUser:         os.Getenv("POSTGRESUSER"),
			PostgresPassword:     os.Getenv("POSTGRESPASSWORD"),
			PostgresAddress:      os.Getenv("POSTGRESADDRESS"),
			PostgresPort:         os.Getenv("POSTGRESPORT"),
			PostgresDatabaseName: os.Getenv("POSTGRESDATABASENAME"),
		},
		Producer: struct {
			ProcuerAddress string
			ProcuerPort    string
		}{
			ProcuerAddress: os.Getenv("PROCUERADDRESS"),
			ProcuerPort:    os.Getenv("PROCUERPORT"),
		},
	}
}

func (c *Config) GetRedisUrl() string {
	return fmt.Sprintf("%s:%s", c.Redis.RedisAddress, c.Redis.RedisPort)
}

func (c *Config) GetPostgresUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.Postgres.PostgresUser, c.Postgres.PostgresPassword, c.Postgres.PostgresAddress, c.Postgres.PostgresPort, c.Postgres.PostgresDatabaseName)
}

func (c *Config) GetProducerUrl() string {
	return fmt.Sprint("%s, %s", c.Producer.ProcuerAddress, c.Producer.ProcuerPort)
}
