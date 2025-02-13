package transaction_audit_insert

import (
	"log/slog"
	"sync"
	"time"
)

const (
	auditInsertQuery = `
INSERT INTO transaction_audit_log (transaction_id, payload, version, event_time) VALUES (?, ?, ?, ?)`
	auditInitQuery = `SELECT 1`
)

// Example – структура для пакетной вставки аудита транзакций.
type Example struct {
	logger         *slog.Logger
	save           func(len int, query string, data []any) (err error)
	size           int
	saveTimeout    int
	saveTime       time.Time
	mu             sync.RWMutex
	transactionIDs []string
	payloads       []string
	versions       []int32
	eventTimes     []time.Time
}

// NewInsertTable возвращает новый экземпляр Example для пакетной вставки аудита.
func NewInsertTable(logger *slog.Logger) *Example {
	logger = logger.With(slog.String("module", "transaction_audit"))
	return &Example{
		logger: logger,
	}
}

func (table *Example) Init(save func(len int, query string, data []any) (err error), size int, saveTimeout int) {
	table.mu.Lock()
	table.save = save
	table.size = size
	table.saveTimeout = saveTimeout
	table.transactionIDs = make([]string, 0, size)
	table.payloads = make([]string, 0, size)
	table.versions = make([]int32, 0, size)
	table.eventTimes = make([]time.Time, 0, size)
	table.mu.Unlock()
}

func (table *Example) QueryInit() string   { return auditInitQuery }
func (table *Example) QueryInsert() string { return auditInsertQuery }

func (table *Example) add(txnID string, payload string, version int, eventTime time.Time) {
	table.mu.Lock()
	table.saveTime = time.Now()
	table.transactionIDs = append(table.transactionIDs, txnID)
	table.payloads = append(table.payloads, payload)
	table.versions = append(table.versions, int32(version))
	table.eventTimes = append(table.eventTimes, eventTime)
	table.mu.Unlock()
}

func (table *Example) Len() (l int) {
	table.mu.RLock()
	l = len(table.transactionIDs)
	table.mu.RUnlock()
	return
}

func (table *Example) Clean() {
	table.mu.Lock()
	table.transactionIDs = append(table.transactionIDs[:0], table.transactionIDs[len(table.transactionIDs):]...)
	table.payloads = append(table.payloads[:0], table.payloads[len(table.payloads):]...)
	table.versions = append(table.versions[:0], table.versions[len(table.versions):]...)
	table.eventTimes = append(table.eventTimes[:0], table.eventTimes[len(table.eventTimes):]...)
	table.mu.Unlock()
}

func (table *Example) Data() (res []any) {
	if table.Len() < 1 {
		return res
	}
	table.mu.RLock()
	res = append(res, table.transactionIDs)
	res = append(res, table.payloads)
	res = append(res, table.versions)
	res = append(res, table.eventTimes)
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

func (table *Example) SaveAudit(txnID string, payload string, version int, eventtime time.Time) (err error) {
	table.add(txnID, payload, version, eventtime)
	if table.Len() >= table.size || time.Since(table.getSaveTime()) > time.Second*time.Duration(table.saveTimeout) {
		err = table.save(table.Len(), table.QueryInsert(), table.Data())
		if err != nil {
			table.logger.Error("failed to save transaction audit log", "error", err.Error(), "txn_id", txnID)
			return
		}
		table.setSaveTime(time.Now())
		table.Clean()
	}
	return
}
