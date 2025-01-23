package env

import (
	"errors"
	"os"
)

type storage interface {
	Env() string
	Config() ([]byte, error)
}

type Module struct {
	storages []storage
}

func New(storages ...storage) *Module {
	return &Module{storages: storages}
}

func (m Module) GetEnv() (string, error) {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		return "local", nil
	}

	return env, nil
}

func (m Module) Init() (interface {
	Config() ([]byte, error)
}, error) {
	env, err := m.GetEnv()
	if err != nil {
		return nil, err
	}
	for i := range m.storages {
		if m.storages[i].Env() == env {
			return m.storages[i], nil
		}
	}
	return nil, errors.New("incorrect ENV")
}
