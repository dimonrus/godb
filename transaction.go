package godb

import (
	"github.com/dimonrus/gohelp"
)

// Generate transaction id
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

// Unset transaction
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

// Transaction count
func (p *TransactionPool) Count() int {
	return len(p.transactions)
}

// Create Transaction pool
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		transactions: make(map[TransactionId]*SqlTx),
	}
}
