package ledger

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type MigrationRepository interface {
	InitializeMigrationsTable(ctx context.Context, db, table string) error
	Migration(ctx context.Context, name string) (*Migration, error)
	MigrationTx(ctx context.Context, tx *sqlx.Tx, name string) (*Migration, error)
	CreateMigration(ctx context.Context, migration *Migration) (*Migration, error)
	CreateMigrationTx(ctx context.Context, tx *sqlx.Tx, migration *Migration) (*Migration, error)
	UpdateMigrationTx(ctx context.Context, tx *sqlx.Tx, migration *Migration) (*Migration, error)
}

type Migration struct {
	ID        uint      `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Executed  bool      `db:"executed" json:"executed"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
