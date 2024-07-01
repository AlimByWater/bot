package test

import (
	"os"
)

const env = "test"

type Storage struct{}

func New() Storage {
	return Storage{}
}

func (Storage) Env() string {
	return env
}

func (Storage) Config() (config []byte, err error) {
	config, err = os.ReadFile("/Users/admin/go/src/arimadj-helper/configs/test.json")
	return
}
