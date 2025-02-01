package click_table_query

import (
	"context"
	"errors"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"time"
)

const (
	exampleSelect = `
select date_create, sum from example order by date_create`
)

type Example struct {
	conn driver.Conn
}

func NewExampleTable() *Example {
	return &Example{}
}

func (table *Example) AddConn(conn driver.Conn) {
	table.conn = conn
}

func (table *Example) GetSum() (dates []time.Time, sums []int64, errs error) {
	ctx := context.Background()
	if errs = table.conn.Ping(ctx); errs != nil {
		return
	}
	rows, errs := table.conn.Query(ctx, exampleSelect)
	if errs != nil {
		return
	}
	for rows.Next() {
		var date time.Time
		var sum int64
		err := rows.Scan(&date, &sum)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		dates = append(dates, date)
		sums = append(sums, sum)
	}
	return
}
