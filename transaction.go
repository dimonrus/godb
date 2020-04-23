package godb

import "github.com/dimonrus/gohelp"

// Generate transaction id
func GenTransactionId() TransactionId {
	return TransactionId(gohelp.RandString(16))
}
