// fb_task
package kfb

import (
	//"container/list"
	"context"
	"database/sql"

	//"fmt"
	//"strconv"
	//"strings"
	//"sync"
	"time"
	//"kot_common/kerr"
	//"kot_common/ksess"
	//"kot_common/kutils"
)

type KfbTask struct {
	// It is "Query","QueryRow" or "Exec"
	queryKind string

	//
	result interface{}
}

func (kfbTask KfbTask) Query(tx *sql.Tx, ctx context.Context, query string, params ...interface{}) {
	if kfbTask.queryKind != "" {
		panic("KfbTask.Query: queryKind is not empty")
	}
	if kfbTask.result != nil {
		panic("KfbTask.Query: result is not nil")
	}

	kfbTask.queryKind = "Query"
	var queryChan = launcheQuery(tx, ctx, query, params)
	kfbTask.result = <-queryChan
}

type QueryResult struct {
	Err     error
	Rows    *sql.Rows
	Elapsed time.Duration
}
type QueryChan chan QueryResult

func launcheQuery(tx *sql.Tx, ctx context.Context, query string, params ...interface{}) (queryChan QueryChan) {
	var queryResult = QueryResult{}
	queryChan = make(chan QueryResult)
	go func() {
		start := time.Now()
		rows, err := tx.QueryContext(ctx, query, params)
		queryResult.Rows = rows
		queryResult.Err = err
		queryResult.Elapsed = time.Since(start)
		queryChan <- queryResult
	}()
	return
}
