package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DiscordToken string
	DBPath       string
}

func Load() (Config, error) {
	// Load .env from the directory of the executable (ignore errors — file may not exist)
	if exe, err := os.Executable(); err == nil {
		loadDotEnv(filepath.Join(filepath.Dir(exe), ".env"))
	}

	var cfg Config
	flag.StringVar(&cfg.DiscordToken, "discord-token", os.Getenv("DISCORD_TOKEN"), "Discord bot token")
	flag.StringVar(&cfg.DBPath, "db-path", getEnv("DB_PATH", "killbot.db"), "Path to SQLite database file")
	flag.Parse()

	if cfg.DiscordToken == "" {
		return cfg, fmt.Errorf("discord token is required (set DISCORD_TOKEN in .env or pass -discord-token)")
	}
	return cfg, nil
}

// loadDotEnv reads KEY=VALUE pairs from path and sets them as env vars
// if they are not already set. Ignores comments and blank lines.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Strip surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		// Don't overwrite values already set in the environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
