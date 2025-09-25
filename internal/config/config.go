package config

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/mitchellh/mapstructure"
    "github.com/spf13/viper"
    "github.com/subosito/gotenv"
)

type Config struct {
    Database DatabaseConfig `mapstructure:"db"`
}

type DatabaseConfig struct {
    // URL overrides individual fields when set (e.g., postgres://user:pass@host:5432/db?sslmode=disable)
    URL            string        `mapstructure:"url"`
    Host           string        `mapstructure:"host"`
    Port           int           `mapstructure:"port"`
    User           string        `mapstructure:"user"`
    Password       string        `mapstructure:"password"`
    Name           string        `mapstructure:"name"`
    SSLMode        string        `mapstructure:"sslmode"` // disable, require, verify-ca, verify-full
    MaxOpenConns   int           `mapstructure:"max_open_conns"`
    MaxIdleConns   int           `mapstructure:"max_idle_conns"`
    ConnMaxLife    time.Duration `mapstructure:"conn_max_lifetime"`
    ConnMaxIdle    time.Duration `mapstructure:"conn_max_idle_time"`
}

// DSN returns a PostgreSQL DSN suitable for pgx stdlib driver.
func (d DatabaseConfig) DSN() string {
    if d.URL != "" {
        return d.URL
    }
    pwd := urlEscape(d.Password)
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", d.User, pwd, d.Host, d.Port, d.Name, d.SSLMode)
}

func urlEscape(s string) string {
    r := strings.NewReplacer("@", "%40", ":", "%3A", "/", "%2F", "?", "%3F", "#", "%23", " ", "%20")
    return r.Replace(s)
}

// Load reads configuration from .env, config file, and environment variables.
// Precedence: defaults < file(s) < env (FLIGHT_*) < explicit env fallbacks (DB_*).
func Load() (*Config, error) {
    // Load .env if present (no error if missing)
    _ = gotenv.Load()

    v := viper.New()
    v.SetEnvPrefix("FLIGHT")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
    v.AutomaticEnv()

    // Defaults
    v.SetDefault("db.host", "localhost")
    v.SetDefault("db.port", 5432)
    v.SetDefault("db.user", "flight_app")
    v.SetDefault("db.password", "app")
    v.SetDefault("db.name", "flight")
    v.SetDefault("db.sslmode", "disable")
    v.SetDefault("db.max_open_conns", 20)
    v.SetDefault("db.max_idle_conns", 10)
    v.SetDefault("db.conn_max_lifetime", "30m")
    v.SetDefault("db.conn_max_idle_time", "5m")

    // Config file discovery: flag may set it externally (root.go), otherwise search
    if v.ConfigFileUsed() == "" {
        if path := os.Getenv("FLIGHT_CONFIG"); path != "" {
            v.SetConfigFile(path)
        } else {
            v.SetConfigName("config")
            v.SetConfigType("yaml")
            v.AddConfigPath(".")
            v.AddConfigPath("./configs")
            v.AddConfigPath("./docker")
        }
        _ = v.ReadInConfig() // ignore if not found
    }

    // Backward-compatible env fallbacks without prefix
    bindEnvAlias(v, "db.host", "DB_HOST")
    bindEnvAlias(v, "db.port", "DB_PORT")
    bindEnvAlias(v, "db.user", "DB_USER")
    bindEnvAlias(v, "db.password", "DB_PASSWORD")
    bindEnvAlias(v, "db.name", "DB_NAME")
    bindEnvAlias(v, "db.sslmode", "DB_SSLMODE")
    bindEnvAlias(v, "db.url", "DATABASE_URL")

    var cfg Config
    dec := func(c *mapstructure.DecoderConfig) {
        c.TagName = "mapstructure"
        c.DecodeHook = mapstructure.StringToTimeDurationHookFunc()
    }
    if err := v.Unmarshal(&cfg, dec); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }
    if err := validate(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

func bindEnvAlias(v *viper.Viper, key, env string) {
    _ = v.BindEnv(key, fmt.Sprintf("FLIGHT_%s", env))
    if _, ok := os.LookupEnv(env); ok {
        _ = v.BindEnv(key, env)
    }
}

func validate(c *Config) error {
    if c.Database.Port < 1 || c.Database.Port > 65535 {
        return errors.New("db.port out of range")
    }
    if strings.TrimSpace(c.Database.Host) == "" {
        return errors.New("db.host required")
    }
    switch c.Database.SSLMode {
    case "", "disable", "require", "verify-ca", "verify-full":
    default:
        return fmt.Errorf("db.sslmode invalid: %s", c.Database.SSLMode)
    }
    return nil
}

// DiscoverConfig attempts to find a config file path candidates (for help output).
func DiscoverConfig() []string {
    var found []string
    for _, p := range []string{"./config.yaml", "./configs/config.yaml", "./docker/config.yaml"} {
        if _, err := os.Stat(filepath.Clean(p)); err == nil {
            found = append(found, p)
        }
    }
    return found
}
