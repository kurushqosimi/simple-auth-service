package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type (
	Config struct {
		App      `yaml:"app"`
		HTTP     `yaml:"http"`
		Log      `yaml:"logger"`
		PG       `yaml:"postgres"`
		Mailer   `yaml:"mailer"`
		Redis    `yaml:"redis"`
		TokenKey `yaml:"token_key"`
	}

	App struct {
		Name    string `env-required:"true" yaml:"name" env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	HTTP struct {
		Port         string        `env-required:"true" yaml:"port" env:"HTTP_PORT"`
		ReadTimeout  time.Duration `env-required:"true" yaml:"read-timeout" env:"HTTP-READ-TIMEOUT"`
		WriteTimeout time.Duration `env-required:"true" yaml:"write-timeout" env:"HTTP-WRITE-TIMEOUT"`
		CORS         struct {
			AllowedMethods     []string `env-required:"true" yaml:"allowed-methods" env:"HTTP-CORS-ALLOWED-METHODS"`
			AllowedOrigins     []string `env-required:"true" yaml:"allowed-origins"`
			AllowCredentials   bool     `env-required:"true" yaml:"allow-credentials"`
			AllowedHeaders     []string `env-required:"true" yaml:"allowed-headers"`
			OptionsPassthrough bool     `yaml:"options-passthrough"`
			ExposedHeaders     []string `env-required:"true" yaml:"exposed-headers"`
			Debug              bool     `env-required:"true" yaml:"debug"`
		} `yaml:"cors"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"log_level" env:"LOG_LEVEL"`
	}

	PG struct {
		PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		URL     string `env-required:"true" yaml:"db_source" env:"DB_SOURCE"`
	}

	Mailer struct {
		SMTPConfig `yaml:"smtp"`
	}

	SMTPConfig struct {
		SenderMail     string `yaml:"sender_mail" env:"SENDER_MAIL"`
		SenderName     string `yaml:"sender_name" env:"SENDER_NAME"`
		SenderPassword string `yaml:"sender_password" env:"SENDER_PASSWORD"`
		SMTPServer     string `yaml:"smtp_server" env:"SMTP_SERVER"`
		SMTPPort       int    `yaml:"smtp_port" env:"SMTP_PORT"`
	}

	Redis struct {
		Addr     string `env-required:"true" yaml:"addr" env:"REDIS_ADDRESS"`
		Password string `env-required:"true" yaml:"password" env:"REDIS_PASSWORD"`
		DB       int    `yaml:"db" env:"REDIS_DB"`
	}

	TokenKey struct {
		TokenSymmetricKey string `yaml:"token_symmetric_key" env:"TOKEN_SYMMETRIC_KEY"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yaml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
