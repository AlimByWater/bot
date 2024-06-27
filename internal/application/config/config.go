package config

import (
	"encoding/json"
	"errors"
	"reflect"
)

// Module структура конфигов
type Module struct {
	configs []any
}

// New констурктор
func New(configs ...any) *Module {
	return &Module{configs: configs}
}

// Init инициализирует конфиги
func (m Module) Init(st interface{ Config() ([]byte, error) }) error {
	data, err := st.Config()
	if err != nil {
		return err
	}
	modules := make(map[string]any, len(m.configs))
	err = json.Unmarshal(data, &modules)
	if err != nil {
		return err
	}

	for i := range m.configs {
		name := reflect.Indirect(reflect.ValueOf(m.configs[i])).Type().Name()
		mod, ok := modules[name]
		if !ok {
			return errors.New("module configuration not found: " + name)
		}
		b, err := json.Marshal(mod)
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, &m.configs[i])
		if err != nil {
			return err
		}
	}
	return nil
}
