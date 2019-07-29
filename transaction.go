package godb

// Transaction callback
type TxCallback func(tx *SqlTx) (interface{}, error)

// Exec func in transaction
func (dbo *DBO) RunInTransaction(callback TxCallback) (interface{}, error) {
	tx, err := dbo.Begin()
	if err != nil {
		return nil, err
	}

	entity, err := callback(tx)

	if err != nil {
		dbo.Logger.Print("Callback error: " + err.Error())
		defer func() {
			err := tx.Rollback()
			if err != nil {
				dbo.Logger.Print("Rollback error: " + err.Error())
			}
		}()
		return nil, err
	}

	dbErr := tx.Commit()
	if dbErr != nil {
		dbo.Logger.Print("Commit error: " + dbErr.Error())
		defer func() {
			err := tx.Rollback()
			if err != nil {
				dbo.Logger.Print("Rollback error: " + err.Error())
			}
		}()
		return nil, err
	}

	return entity, nil
}
