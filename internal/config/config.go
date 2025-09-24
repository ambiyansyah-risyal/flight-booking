package config

import (
    "fmt"
    "os"
    "strings"

    "github.com/spf13/viper"
)

type Config struct {
    Database DatabaseConfig
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
    SSLMode  string // disable, require, verify-ca, verify-full
}

// DSN returns a PostgreSQL DSN suitable for pgx stdlib driver.
func (d DatabaseConfig) DSN() string {
    // Prefer URL-style DSN
    pwd := urlEscape(d.Password)
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", d.User, pwd, d.Host, d.Port, d.Name, d.SSLMode)
}

func urlEscape(s string) string {
    r := strings.NewReplacer("@", "%40", ":", "%3A", "/", "%2F", "?", "%3F", "#", "%23", " ", "%20")
    return r.Replace(s)
}

// Load reads configuration from environment variables and optional file.
// Environment variables use the FLIGHT_ prefix by convention.
func Load() (*Config, error) {
    viper.SetEnvPrefix("FLIGHT")
    viper.AutomaticEnv()

    viper.SetDefault("db.host", "localhost")
    viper.SetDefault("db.port", 5432)
    viper.SetDefault("db.user", "flight_app")
    viper.SetDefault("db.password", "app")
    viper.SetDefault("db.name", "flight")
    viper.SetDefault("db.sslmode", "disable")

    // Map env overrides
    bindEnvAlias("db.host", "DB_HOST")
    bindEnvAlias("db.port", "DB_PORT")
    bindEnvAlias("db.user", "DB_USER")
    bindEnvAlias("db.password", "DB_PASSWORD")
    bindEnvAlias("db.name", "DB_NAME")
    bindEnvAlias("db.sslmode", "DB_SSLMODE")

    cfg := &Config{
        Database: DatabaseConfig{
            Host:     viper.GetString("db.host"),
            Port:     viper.GetInt("db.port"),
            User:     viper.GetString("db.user"),
            Password: viper.GetString("db.password"),
            Name:     viper.GetString("db.name"),
            SSLMode:  viper.GetString("db.sslmode"),
        },
    }
    return cfg, nil
}

func bindEnvAlias(key, env string) {
    // Map FLIGHT_<env> to key (e.g., FLIGHT_DB_HOST)
    _ = viper.BindEnv(key, fmt.Sprintf("FLIGHT_%s", env))
    // Fallback without prefix if explicitly set in environment
    if _, ok := os.LookupEnv(env); ok {
        _ = viper.BindEnv(key, env)
    }
}

