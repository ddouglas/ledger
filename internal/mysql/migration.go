package mysql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type migrationRepository struct {
	db *sqlx.DB
}

const createMigrationTableQuery = `
	CREATE TABLE migrations (
		id int unsigned not null auto_increment,
		name varchar(255) not null,
		executed tinyint(1) not null default 0,
		created_at datetime not null,
		updated_at datetime not null,
		primary key (id) using btree,
		unique index migrations_name_unique_idx (name) using btree
	) COLLATE = 'utf8mb4_bin' ENGINE = INNODB;
`

func NewMigrationRepostory(db *sqlx.DB) ledger.MigrationRepository {
	return &migrationRepository{db}
}

func (r *migrationRepository) InitializeMigrationsTable(ctx context.Context, db, table string) error {

	query, args, err := sq.Select("COUNT(*) as count").From("information_schema.tables").Where(sq.Eq{
		"table_schema": db,
		"table_name":   table,
	}).Limit(1).ToSql()
	if err != nil {
		return errors.Wrap(err, "[InitializeMigrationsTable] failed to generate query")
	}

	var count int
	err = r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return errors.Wrap(err, "[InitializeMigrationsTable] failed to if migration table exists")
	}

	if count > 0 {
		return nil
	}

	_, err = r.db.ExecContext(ctx, createMigrationTableQuery)
	return errors.Wrap(err, "[InitializeMigrationsTable] failed to create migrations table")

}

func (r *migrationRepository) Migration(ctx context.Context, name string) (*ledger.Migration, error) {

	query, args, err := sq.Select("id", "name", "created_at").From("migrations").Where(sq.Eq{
		"name": name,
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[Migration] failed to generate query")
	}

	var migration = new(ledger.Migration)
	err = r.db.GetContext(ctx, migration, query, args...)
	return migration, errors.Wrap(err, "[Migration] failed to fetch migration")

}

func (r *migrationRepository) MigrationTx(ctx context.Context, tx *sqlx.Tx, name string) (*ledger.Migration, error) {

	query, args, err := sq.Select("id", "name", "created_at").From("migrations").Where(sq.Eq{
		"name": name,
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[MigrationTx] failed to generate query")
	}

	var migration = new(ledger.Migration)
	err = tx.GetContext(ctx, migration, query, args...)
	return migration, errors.Wrap(err, "[MigrationTx] failed to fetch migration")

}

func (r *migrationRepository) CreateMigration(ctx context.Context, migration *ledger.Migration) (*ledger.Migration, error) {

	query, args, err := sq.Insert("migrations").SetMap(map[string]interface{}{
		"name":       migration.Name,
		"executed":   migration.Executed,
		"created_at": sq.Expr(`NOW()`),
		"updated_at": sq.Expr(`NOW()`),
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[CreateMigration] failed to generate query")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[CreateMigration] failed to create migration")
	}

	return r.Migration(ctx, migration.Name)

}

func (r *migrationRepository) CreateMigrationTx(ctx context.Context, tx *sqlx.Tx, migration *ledger.Migration) (*ledger.Migration, error) {

	query, args, err := sq.Insert("migrations").SetMap(map[string]interface{}{
		"name":       migration.Name,
		"executed":   migration.Executed,
		"created_at": sq.Expr(`NOW()`),
		"updated_at": sq.Expr(`NOW()`),
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[CreateMigrationTx] failed to generate query")
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[CreateMigrationTx] failed to create migration")
	}

	return r.MigrationTx(ctx, tx, migration.Name)

}

func (r *migrationRepository) UpdateMigrationTx(ctx context.Context, tx *sqlx.Tx, migration *ledger.Migration) (*ledger.Migration, error) {

	query, args, err := sq.Update("migrations").SetMap(map[string]interface{}{
		"executed":   migration.Executed,
		"updated_at": sq.Expr(`NOW()`),
	}).Where(sq.Eq{
		"name": migration.Name,
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[UpdateMigrationTx] failed to generate query")
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[UpdateMigrationTx] failed to update migration")
	}

	return r.MigrationTx(ctx, tx, migration.Name)

}
