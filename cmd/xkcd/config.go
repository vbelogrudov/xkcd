package main

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	DatabasePath string `yaml:"db_path" env:"DB" env-default:"database.json"`
	SourceURL    string `yaml:"source_url" env:"URL" env-default:"https://xkcd.com"`
	Parallel     int    `jaml:"parallel" env:"PARALLEL" env-default:"32"`
}

func GetConfig(path string) (Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig(path, &cfg)
	return cfg, err
}
