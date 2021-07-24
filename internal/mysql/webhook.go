package mysql

import (
	"context"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type webhookRepository struct {
	db *sqlx.DB
}

var webhookLogTableName = "webhook_log"

var webhookLogColumns = []string{
	"id", "payload", "created_at",
}

func NewWebhookRepository(db *sqlx.DB) ledger.WebhookRepository {
	return &webhookRepository{
		db,
	}
}

func (r *webhookRepository) LogWebhook(ctx context.Context, webhook *ledger.WebhookMessage) error {

	columns := webhookLogColumns[:1]

	data, err := json.Marshal(webhook)
	if err != nil {
		return errors.Wrap(err, "[mysql.LogWebhook]")
	}

	query, args, err := sq.Insert(webhookLogTableName).Columns(columns...).Values(
		data, sq.Expr(`NOW()`),
	).ToSql()
	if err != nil {
		return errors.Wrap(err, "[mysql.LogWebhook]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)

	return errors.Wrap(err, "[mysql.LogWebhook]")

}
