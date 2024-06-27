package local

import (
	"os"
)

const env = "local"

type Storage struct{}

func New() Storage {
	return Storage{}
}

func (Storage) Env() string {
	return env
}

func (Storage) Config() (config []byte, err error) {
	config, err = os.ReadFile("./configs/local.json")
	return
}
