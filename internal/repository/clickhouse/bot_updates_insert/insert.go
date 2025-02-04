package bot_updates_insert

import (
	"elysium/internal/entity"
	"sync"
	"time"
)

const (
	exampleInsert = `
INSERT INTO bot_updates (bot_id, update_time, payload) VALUES (?, ?, ?)`

	exampleInit = `SELECT 1`
)

type Example struct {
	save        func(len int, query string, data []any) (err error)
	size        int
	saveTimeout int
	saveTime    time.Time
	mu          sync.RWMutex
	botIds      []int64
	updateTimes []time.Time
	payloads    []string
}

func NewInsertTable() *Example {
	return &Example{}
}

func (table *Example) Init(save func(len int, query string, data []any) (err error), size int, saveTimeout int) {
	table.mu.Lock()
	table.save = save
	table.size = size
	table.saveTimeout = saveTimeout
	table.botIds = make([]int64, 0, size)
	table.updateTimes = make([]time.Time, 0, size)
	table.payloads = make([]string, 0, size)
	table.mu.Unlock()
}

func (table *Example) QueryInit() string   { return exampleInit }
func (table *Example) QueryInsert() string { return exampleInsert }

func (table *Example) add(botUpdate entity.BotUpdate) {
	table.mu.Lock()
	table.botIds = append(table.botIds, botUpdate.BotID)
	table.updateTimes = append(table.updateTimes, botUpdate.UpdateTime)
	table.payloads = append(table.payloads, botUpdate.Payload)
	table.mu.Unlock()
}

func (table *Example) Len() (l int) {
	table.mu.RLock()
	l = len(table.payloads)
	table.mu.RUnlock()
	return
}

func (table *Example) Clean() {
	table.mu.Lock()
	table.botIds = append(table.botIds[:0], table.botIds[len(table.botIds):]...)
	table.updateTimes = append(table.updateTimes[:0], table.updateTimes[len(table.updateTimes):]...)
	table.payloads = append(table.payloads[:0], table.payloads[len(table.payloads):]...)
	table.mu.Unlock()
}

func (table *Example) Data() (res []any) {
	if table.Len() < 1 {
		return res
	}
	table.mu.RLock()
	res = append(res, table.botIds)
	res = append(res, table.updateTimes)
	res = append(res, table.payloads)
	table.mu.RUnlock()
	return res
}

func (table *Example) getSaveTime() (t time.Time) {
	table.mu.RLock()
	t = table.saveTime
	table.mu.RUnlock()
	return
}

func (table *Example) setSaveTime(t time.Time) {
	table.mu.Lock()
	table.saveTime = t
	table.mu.Unlock()
}

func (table *Example) SaveUpdate(botUpdate entity.BotUpdate) (err error) {
	table.add(botUpdate)
	//fmt.Println(table.Len(), table.size, time.Since(table.getSaveTime()), time.Second*time.Duration(table.saveTimeout))
	if table.Len() >= table.size || time.Since(table.getSaveTime()) > time.Second*time.Duration(table.saveTimeout) {
		err = table.save(table.Len(), table.QueryInsert(), table.Data())
		if err != nil {
			return
		}
		table.setSaveTime(time.Now())
		table.Clean()
	}
	return
}
