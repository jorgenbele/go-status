package status

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"time"
)

type FsNotifyTicker struct {
	C    <-chan time.Time
	stop chan bool
}

// Stop the FsNotifyTicker
func (t *FsNotifyTicker) Stop() {
	t.stop <- true
}
// NewFsNotifyTicker creates ticker which ticks for each fsnotify on
// one of the provided paths.
func NewFsNotifyTicker(paths []string) (ticker FsNotifyTicker) {
	// File watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan time.Time)
	ticker.C = c
	ticker.stop = make(chan bool)

	for _, p := range paths {
		err = watcher.Add(p)
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ticker.stop:
				log.Printf("Stopping FsNotifyTick on paths: %v\n", paths)
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("FsNotifyTick event:", event)
				c <- time.Now()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	return
}
