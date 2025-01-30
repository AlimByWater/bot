package use_message

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"reflect"
)

type config interface {
	GetDefaultLanguage() string
	GetDirectoryName() string
}

type language struct {
	Code          string
	Error         string
	BalanceBtn    string
	BotsListBtn   string
	SupportBtn    string
	BuyTokens     string
	BuyTokensBtn  string
	StartDripTech string
	BackBtn       string
}

type Module struct {
	logger    *slog.Logger
	config    config
	languages []language
	defIdx    int
}

func New(config config) *Module {
	return &Module{
		config: config,
	}
}

func (m *Module) Init(_ context.Context, logger *slog.Logger) (err error) {
	m.logger = logger.With(slog.String("module", reflect.Indirect(reflect.ValueOf(m)).Type().PkgPath()))

	dir, err := os.ReadDir(m.config.GetDirectoryName())
	if err != nil {
		return
	}
	var defIdxOk bool
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		err = func() error {
			file, err := os.Open(m.config.GetDirectoryName() + entry.Name())
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close()
			}()

			info, err := entry.Info()
			if err != nil {
				return err
			}
			buf := make([]byte, info.Size())
			i, err := file.Read(buf)
			if err != nil {
				return err
			}
			buf = buf[:i]
			var l language
			err = json.Unmarshal(buf, &l)
			if err != nil {
				return err
			}
			m.languages = append(m.languages, l)
			if l.Code == m.config.GetDefaultLanguage() {
				m.defIdx = len(m.languages) - 1
				defIdxOk = true
			}
			return nil
		}()
		if err != nil {
			return
		}
	}
	if !defIdxOk {
		err = errors.New("no default lang code: " + m.config.GetDefaultLanguage())
	}
	return
}

func (m *Module) langIdx(langCode string) (idx int) {
	for i := range m.languages {
		if m.languages[i].Code == langCode {
			return i
		}
	}
	return m.defIdx
}

func (m *Module) Error(langCode string) (msg string) {
	return m.languages[m.langIdx(langCode)].Error
}

func (m *Module) BalanceBtn(langCode string) string {
	return m.languages[m.langIdx(langCode)].BalanceBtn
}

func (m *Module) BotsListBtn(langCode string) string {
	return m.languages[m.langIdx(langCode)].BotsListBtn
}

func (m *Module) SupportBtn(langCode string) string {
	return m.languages[m.langIdx(langCode)].SupportBtn
}

func (m *Module) BuyTokens(langCode string) string {
	return m.languages[m.langIdx(langCode)].BuyTokens
}

func (m *Module) BuyTokensBtn(langCode string) string {
	return m.languages[m.langIdx(langCode)].BuyTokensBtn
}

func (m *Module) StartDripTech(langCode string) string {
	return m.languages[m.langIdx(langCode)].StartDripTech
}

func (m *Module) BackBtn(langCode string) string {
	return m.languages[m.langIdx(langCode)].BackBtn
}
