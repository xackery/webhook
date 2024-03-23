package config

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/jbsmith7741/toml"
	"github.com/rs/zerolog"
)

// Config represents a configuration parse
type Config struct {
	Debug  bool    `toml:"debug" desc:"webhook Configuration\n\n# Debug messages are displayed. This will cause console to be more verbose, but also more informative"`
	Events []Event `toml:"events" desc:"Events to listen for"`
}

type Event struct {
	Name           string   `toml:"name" desc:"Event name"`
	WebhookToken   string   `toml:"webhook_token" desc:"Webhook token"`
	DiscordWebhook string   `toml:"discord_webhook" desc:"Discord webhook"`
	Path           string   `toml:"path" desc:"Working directory path for hook trigger"`
	Command        string   `toml:"command" desc:"Command to run when hook is triggered"`
	Args           []string `toml:"args" desc:"Arguments to pass to command when hook is triggered"`
}

// NewConfig creates a new configuration
func NewConfig(ctx context.Context) (*Config, error) {
	var f *os.File
	cfg := Config{}
	path := "webhook.conf"

	isNewConfig := false
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("config info: %w", err)
		}
		f, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("create webhook.conf: %w", err)
		}
		fi, err = os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("new config info: %w", err)
		}
		isNewConfig = true
	}
	if !isNewConfig {
		f, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open config: %w", err)
		}
	}

	defer f.Close()
	if fi.IsDir() {
		return nil, fmt.Errorf("webhook.conf is a directory, should be a file")
	}

	if isNewConfig {
		enc := toml.NewEncoder(f)
		enc.Encode(getDefaultConfig())

		fmt.Println("a new webhook.conf file was created. Please open this file and configure webhook, then run it again.")
		if runtime.GOOS == "windows" {
			option := ""
			fmt.Println("press a key then enter to exit.")
			fmt.Scan(&option)
		}
		os.Exit(0)
	}

	_, err = toml.DecodeReader(f, &cfg)
	if err != nil {
		return nil, fmt.Errorf("decode webhook.conf: %w", err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	err = cfg.Verify()
	if err != nil {
		return nil, fmt.Errorf("verify: %w", err)
	}

	return &cfg, nil
}

// Verify returns an error if configuration appears off
func (c *Config) Verify() error {

	return nil
}

func getDefaultConfig() Config {
	cfg := Config{
		Debug: true,
	}

	return cfg
}
