package processing

import (
	"log/slog"
	"sync"
)

type Module struct {
	directories sync.Map
	logger      *slog.Logger
}

func NewProcessingModule(logger *slog.Logger) *Module {
	return &Module{
		directories: sync.Map{},
		logger:      logger,
	}
}
