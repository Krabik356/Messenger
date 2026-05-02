package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server struct {
		ServerAddres string
		ServerPort   string
	}
	Consumer struct {
		ConsumerAddres string
		ConsumerPort   string
	}
	Postgres struct {
		PostgresUser         string
		PostgresPassword     string
		PostgresAddress      string
		PostgresPort         string
		PostgresDatabaseName string
	}
	TokensSecret string
}

func NewConfig() *Config {
	return &Config{
		Server: struct {
			ServerAddres string
			ServerPort   string
		}{
			ServerAddres: os.Getenv("SERVERADDRES"),
			ServerPort:   os.Getenv("SERVERPORT"),
		},
		Consumer: struct {
			ConsumerAddres string
			ConsumerPort   string
		}{
			ConsumerAddres: os.Getenv("CONSUMERADDRES"),
			ConsumerPort:   os.Getenv("CONSUMERPORT"),
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
		TokensSecret: os.Getenv("TOKENSSECRET"),
	}
}

func (c *Config) GetPostgresUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.Postgres.PostgresUser, c.Postgres.PostgresPassword, c.Postgres.PostgresAddress, c.Postgres.PostgresPort, c.Postgres.PostgresDatabaseName)
}

func (c *Config) GetServerUrl() string {
	return fmt.Sprintf("%s:%s", c.Server.ServerAddres, c.Server.ServerPort)
}

func (c *Config) GetConsumerUrl() string {
	return fmt.Sprintf("%s:%s", c.Consumer.ConsumerAddres, c.Consumer.ConsumerPort)
}
