package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"example.com/xkcd/pkg/app"
	"example.com/xkcd/pkg/database"
	"example.com/xkcd/pkg/xkcd"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "./config.yaml", "path to config file")
	flag.Parse()

	cfg, err := GetConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	db := database.New(cfg.DatabasePath)
	cl := xkcd.New(cfg.SourceURL)
	app := app.New(cl, db, cfg.Parallel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err = app.Sync(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
