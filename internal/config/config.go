package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	PgUser           string
	PgPassword       string
	PgHost           string
	PgPort           string
	PgDatabase       string
	PgSslMode        string
	GOauthClientId   string
	HttpMaxBodyBytes int64
}

func Init() *Config {
	c := &Config{}
	c.loadConfig()
	return c
}

func (c *Config) loadConfig() {
	c.PgUser = getEnvStr("PGUSER")
	c.PgPassword = getEnvStr("PGPASSWORD")
	c.PgHost = getEnvStr("PGHOST")
	c.PgPort = getEnvStr("PGPORT")
	c.PgDatabase = getEnvStr("PGDATABASE")
	c.PgSslMode = getEnvStr("PGSSLMODE") // disable
	c.HttpMaxBodyBytes = 1024 * 1024
	c.GOauthClientId = getEnvStr("GOAUTHCLIENTID")
}

func (c *Config) GetPgDsn() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		c.PgUser, c.PgPassword, c.PgHost, c.PgPort, c.PgDatabase, c.PgSslMode)
}

func getEnvStr(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Panicf("Error loading Config: you must set '%s' Environment Variable", key)
	}
	return value
}
