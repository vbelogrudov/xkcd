package app

import (
	"context"
	"errors"
	"log"
	"time"

	"example.com/xkcd/pkg/database"
	"example.com/xkcd/pkg/xkcd"
)

type App struct {
	client   *xkcd.XKCD
	database *database.DB
	parallel int
}

func New(cl *xkcd.XKCD, db *database.DB, parallel int) *App {
	return &App{
		client:   cl,
		database: db,
		parallel: parallel,
	}
}

func (a *App) Sync(ctx context.Context) (err error) {
	defer func(start time.Time) {
		log.Printf("finished sync in %v", time.Since(start))
	}(time.Now())
	log.Println("syncing with XKCD site")
	// always save on exit
	defer func() {
		dbErr := a.database.Save()
		err = errors.Join(err, dbErr)
	}()

	return a.runSyncPipeline(ctx)
}
