// asynch
//The new functionality of 191102 version
package kpglo

import (
	"fmt"
	//"io"
	"time"

	"github.com/go-pg/pg"
)

const MinChunckSizeForAsynchRead = 100000

type ReadChunkRep struct {
	//Loid int
	Err error
	//Beg        time.Time
	Dur       time.Duration
	ChunkData []byte //if it = nil and Err!=nil than reading of all chunks is over successfully
}

//AsynchReadLo creates a channel (loChan) and send to it chunks of a LO in form of ReadChunckRep
//That is, into the channel there are succession of chunks, the last of which must have ReadChunckRep.Err!=nil or ReadChunckRep.Err==nil and ReadChunckRep.Data==nil
//Inside itself the function gets a transaction and does all its work under the transaction
//The transaction is rollbacked in any case, that is the database is returned to its initial state independently of result
func AsynchReadLo(loid int, chunkSize int) (chunkChan chan ReadChunkRep) {
	var (
		buff []byte
		err  error
		//rep     ReadChunkRep
		tx *pg.Tx
		//start   time.Time
		loDescr int
	)
	start := time.Now()

	if chunkSize < MinChunckSizeForAsynchRead {
		chunkSize = MinChunckSizeForAsynchRead
	}

	chunkChan = make(chan ReadChunkRep)

	if tx, err = GetPGTx(); err != nil {
		err = fmt.Errorf("AsynchReadLo: GetPGTx() err = %v", err.Error())
		rep := ReadChunkRep{err, time.Since(start), nil}
		chunkChan <- rep
		return
	}
	if loDescr, err = openLo(tx, int64(loid), 0x00040000); err != nil {
		err = fmt.Errorf("AsynchReadLo:  openLo err = %v", err.Error())
		rep := ReadChunkRep{err, time.Since(start), nil}
		tx.Rollback()
		chunkChan <- rep
		return
	}

	go func() {
		start := time.Now()
		if buff, err = readFromLo(tx, loDescr, chunkSize); err != nil {
			err = fmt.Errorf("AsynchReadLo:  ReadFromLo err = %v", err.Error())
			rep := ReadChunkRep{err, time.Since(start), nil}
			chunkChan <- rep
			tx.Rollback()
			return
		} else {
			rep := ReadChunkRep{nil, time.Since(start), buff}
			chunkChan <- rep

		}
		for len(buff) != 0 {
			start := time.Now()
			if buff, err = readFromLo(tx, loDescr, chunkSize); err != nil {
				err = fmt.Errorf("AsynchReadLo:  ReadFromLo err = %v", err.Error())
				rep := ReadChunkRep{err, time.Since(start), nil}
				chunkChan <- rep
				tx.Rollback()
				return
			} else {
				rep := ReadChunkRep{nil, time.Since(start), buff}
				chunkChan <- rep
			}
		}
		tx.Rollback()
	}() //go
	return
}
