package ledger

type Starter interface {
	Begin() (Transactioner, error)
}

type Transactioner interface {
	Commit() error
	Rollback() error
}
