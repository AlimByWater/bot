package scheduler

import (
	"context"
	"github.com/go-co-op/gocron"
	"log/slog"
	"reflect"
	"time"
)

type job interface {
	Do(*gocron.Scheduler) (*gocron.Job, error)
}

func New(reconfiguration chan struct{}, jobs ...job) *Module {
	return &Module{jobs: jobs, reconfiguration: reconfiguration}
}

type Module struct {
	ctx             context.Context
	stop            context.CancelFunc
	logger          *slog.Logger
	scheduler       *gocron.Scheduler
	reconfiguration chan struct{}
	jobs            []job
}

func (m *Module) Init(ctx context.Context, stop context.CancelFunc, logger *slog.Logger) (err error) {
	m.ctx = ctx
	m.stop = stop
	m.logger = logger.With(slog.String("module", reflect.Indirect(reflect.ValueOf(m)).Type().PkgPath()))

	m.scheduler = gocron.NewScheduler(time.UTC)
	m.scheduler.SetMaxConcurrentJobs(0, gocron.RescheduleMode)
	for _, j := range m.jobs {
		_, err = j.Do(m.scheduler)
		if err != nil {
			return
		}
	}
	return
}

func (m *Module) Run() {
	m.scheduler.StartAsync()
	go m.recfg()
}

func (m *Module) recfg() {
	select {
	case <-m.ctx.Done():
		return
	case <-m.reconfiguration:
		_ = m.Shutdown()
		if err := m.Init(m.ctx, m.stop, m.logger); err != nil {
			m.logger.Error(err.Error())
			m.stop()
			return
		}
		m.Run()
	}
}

func (m *Module) Shutdown() (err error) {
	if m.scheduler != nil {
		m.scheduler.Stop()
	}
	return
}
