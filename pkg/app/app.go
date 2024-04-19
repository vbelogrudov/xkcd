package app

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"example.com/xkcd/pkg/database"
	"example.com/xkcd/pkg/words"
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

	return a.runPipeline(ctx)
}

func (a *App) runPipeline(ctx context.Context) error {
	// need to stop pipeline if got more than 2 "not found" errors
	noMoreComicsCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// assemble pipeline
	idStream := a.generateTasks(noMoreComicsCtx)
	filterStream := a.filterTasks(ctx, idStream)
	fanOutStreams := a.fanoutTasks(ctx, filterStream, a.fetchTasks, a.parallel)
	fanInStream := a.faninTasks(ctx, fanOutStreams)
	stemStream := a.stemTasks(ctx, fanInStream)
	saveStream := a.saveTasks(ctx, stemStream)

	var errs []error
	var notFound int
	var processed int
	for t := range saveStream {
		processed++
		if processed%100 == 0 {
			log.Printf("processed %d comics, last processed id %d", processed, t.id)
		}
		if t.err == nil {
			continue
		}
		if errors.Is(t.err, xkcd.ErrNotFound) {
			notFound++
			if notFound > 1 {
				// no more comics
				cancel()
			}
			continue
		}
		errs = append(errs, t.err)
	}

	log.Printf("total number of comics: %d", a.database.Size())
	return errors.Join(errs...)
}

type task struct {
	id       int
	img      string
	phrase   string
	keywords []string
	err      error
}

func (a *App) generateTasks(ctx context.Context) <-chan task {
	output := make(chan task)

	go func() {
		defer close(output)
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				return
			case output <- task{id: i}:
			}
		}
	}()

	return output
}

func (a *App) filterTasks(ctx context.Context, input <-chan task) <-chan task {
	output := make(chan task)
	existingKeys := a.database.Keys()
	existingKeysMap := make(map[int]bool, len(existingKeys))
	for _, id := range existingKeys {
		existingKeysMap[id] = true
	}

	go func() {
		var skipped int
		defer func() {
			log.Printf("skipped %d existing comics", skipped)
		}()

		defer close(output)
		for t := range input {
			if existingKeysMap[t.id] {
				skipped++
				continue
			}
			select {
			case <-ctx.Done():
				return
			case output <- t:
			}
		}
	}()

	return output
}

func (a *App) fanoutTasks(
	ctx context.Context,
	input <-chan task,
	pipeTasks func(context.Context, <-chan task) <-chan task,
	num int,
) []<-chan task {
	outputs := make([]<-chan task, 0, num)
	for range num {
		outputs = append(outputs, pipeTasks(ctx, input))
	}
	return outputs
}

func (a *App) faninTasks(ctx context.Context, inputs []<-chan task) <-chan task {
	output := make(chan task)
	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, input := range inputs {
		input := input // anachronism since 1.22
		go func() {
			defer wg.Done()
			for t := range input {
				select {
				case <-ctx.Done():
					return
				case output <- t:
				}
			}
		}()
	}

	// need one closer goroutine for all
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

func (a *App) fetchTasks(ctx context.Context, input <-chan task) <-chan task {
	output := make(chan task)

	go func() {
		defer close(output)
		for t := range input {
			comics, err := a.client.Get(t.id)
			t.err = err
			t.img = comics.Image
			t.phrase = comics.Title + " " + comics.Transcript + " " + comics.Alt
			select {
			case <-ctx.Done():
				return
			case output <- t:
			}
		}
	}()

	return output
}

func (a *App) stemTasks(ctx context.Context, input <-chan task) <-chan task {
	output := make(chan task)

	go func() {
		defer close(output)
		for t := range input {
			t.keywords = words.Stem(t.phrase)
			select {
			case <-ctx.Done():
				return
			case output <- t:
			}
		}
	}()

	return output
}

func (a *App) saveTasks(ctx context.Context, input <-chan task) <-chan task {
	output := make(chan task)

	go func() {
		defer close(output)
		for t := range input {
			if t.err == nil {
				a.database.Add(database.Comics{
					ID:       t.id,
					URL:      t.img,
					Keywords: t.keywords,
				})
			}
			select {
			case <-ctx.Done():
				return
			case output <- t:
			}
		}
	}()

	return output
}
