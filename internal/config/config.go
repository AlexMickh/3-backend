package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env    string `env:"ENV" env-default:"prod"`
	Server ServerConfig
	DB     DBConfig
	Jwt    JwtConfig
	Tokens TokensConfig
	Mail   MailConfig
}

type ServerConfig struct {
	Addr          string        `env:"SERVER_ADDR" env-default:"0.0.0.0:50070"`
	Timeout       time.Duration `env:"SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout   time.Duration `env:"SERVER_IDLE_TIMEOUT" env-default:"60s"`
	AdminLogin    string        `env:"ADMIN_LOGIN" env-required:"true"`
	AdminPassword string        `env:"ADMIN_PASSWORD" env-required:"true"`
}

type DBConfig struct {
	File string `env:"DB_FILE" env-default:"storage.db"`
}

type JwtConfig struct {
	AccessTokenTtl  time.Duration `env:"JWT_ACCESS_TOKEN_TTL" env-default:"15m"`
	RefreshTokenTtl time.Duration `env:"JWT_REFRESH_TOKEN_TTL" env-default:"43200m"`
	Secret          string        `env:"JWT_SECRET" env-required:"true"`
}

type TokensConfig struct {
	VerifyEmailTokenTtl time.Duration `env:"TOKENS_VERIFY_EMAIL_TOKEN_TTL" env-default:"15m"`
}

type MailConfig struct {
	Host     string `env:"MAIL_HOST" yaml:"host" env-required:"true"`
	Port     int    `env:"MAIL_PORT" yaml:"port" env-required:"true"`
	FromAddr string `env:"MAIL_FROM_ADDR" yaml:"from_addr" env-required:"true"`
	Password string `env:"MAIL_PASSWORD" yaml:"password" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchPath()
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func fetchPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
