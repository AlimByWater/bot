package prod

import (
	"os"
)

const env = "prod"

type Storage struct{}

func New() Storage {
	return Storage{}
}

func (Storage) Env() string {
	return env
}

func (Storage) Config() (config []byte, err error) {
	config, err = os.ReadFile("./configs/prod.json")
	return
}
