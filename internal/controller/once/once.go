package once

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
)

type job interface {
	Init(wg *sync.WaitGroup)
	Do()
}

func New(close bool, jobs ...job) *Module {
	return &Module{close: close, jobs: jobs}
}

type Module struct {
	ctx    context.Context
	stop   context.CancelFunc
	logger *slog.Logger
	jobs   []job
	wg     *sync.WaitGroup
	close  bool
}

func (m *Module) Init(ctx context.Context, stop context.CancelFunc, logger *slog.Logger) (err error) {
	m.wg = &sync.WaitGroup{}
	m.ctx = ctx
	m.stop = stop
	m.logger = logger.With(slog.String("module", reflect.Indirect(reflect.ValueOf(m)).Type().PkgPath()))

	for i := range m.jobs {
		m.jobs[i].Init(m.wg)
	}
	return
}

func (m *Module) Run() {
	go m.run()
}

func (m *Module) run() {
	for i := range m.jobs {
		select {
		case <-m.ctx.Done():
			return
		default:
			m.wg.Add(1)
			go m.jobs[i].Do()
		}
	}
	if m.close {
		m.wg.Wait()
		m.logger.Info("All job completed")
		m.stop()
	}
}

func (m *Module) Shutdown() (err error) {
	m.wg.Wait()
	return
}
