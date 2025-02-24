package async

import (
	"log"
	"runtime/debug"
	"sync"
)

type BackgroundRunner struct {
	wg *sync.WaitGroup
}

func NewBackgroundRunner(wg *sync.WaitGroup) *BackgroundRunner {
	return &BackgroundRunner{wg: wg}
}

func (r *BackgroundRunner) RunAsync(fn func()) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in background: %v\n%s", rec, debug.Stack())
			}
		}()
		fn()
	}()
}
