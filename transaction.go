package godb

import (
	"github.com/dimonrus/gohelp"
	"sync"
)

// TransactionId transaction identifier
type TransactionId string

// TransactionPool transaction pool
type TransactionPool struct {
	transactions map[TransactionId]*SqlTx
	m            sync.RWMutex
}

// Transaction params
type Transaction struct {
	// Time to live in unix timestampt
	// 0 - no TTL for transaction
	TTL int
	// Event on transaction done
	done chan struct{}
}

// GenTransactionId Generate transaction id
func GenTransactionId() TransactionId {
	return TransactionId(gohelp.RandString(16))
}

// Get transaction if exists
func (p *TransactionPool) Get(id TransactionId) *SqlTx {
	p.m.RLock()
	tx := p.transactions[id]
	p.m.RUnlock()
	return tx
}

// Set transaction
func (p *TransactionPool) Set(id TransactionId, tx *SqlTx) *TransactionPool {
	p.m.Lock()
	p.transactions[id] = tx
	p.m.Unlock()
	return p
}

// UnSet transaction
func (p *TransactionPool) UnSet(id TransactionId) *TransactionPool {
	p.m.Lock()
	delete(p.transactions, id)
	p.m.Unlock()
	return p
}

// Reset pool
func (p *TransactionPool) Reset() *TransactionPool {
	p.m.Lock()
	p.transactions = make(map[TransactionId]*SqlTx)
	p.m.Unlock()
	return p
}

// Count transaction count
func (p *TransactionPool) Count() int {
	return len(p.transactions)
}

// NewTransactionPool Create transaction pool
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		transactions: make(map[TransactionId]*SqlTx),
	}
}
