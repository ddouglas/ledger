package ledger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/plaid/plaid-go/plaid"
	"github.com/volatiletech/null"
)

type TransactionRepository interface {
	Transaction(ctx context.Context, itemID, transactionID string) (*Transaction, error)
	TransactionsCount(ctx context.Context, itemID, accountID string, filters *TransactionFilter) (uint64, error)
	TransactionsPaginated(ctx context.Context, itemID, accountID string, filters *TransactionFilter) ([]*Transaction, error)
	CreateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	UpdateTransaction(ctx context.Context, transactionID string, transaction *Transaction) (*Transaction, error)
	UpdateTransactionMerchantTx(ctx context.Context, txn Transactioner, byMerchantID, toMerchantID string) error
}

type PaginatedTransactions struct {
	Transactions []*Transaction `json:"transactions"`
	Total        uint64         `json:"total"`
}

type Transaction struct {
	ItemID                 string      `db:"item_id" json:"itemID" diff:"-"`
	AccountID              string      `db:"account_id" json:"accountID"`
	TransactionID          string      `db:"transaction_id" json:"transactionID"`
	PendingTransactionID   null.String `db:"pending_transaction_id" json:"pendingTransactionID"`
	CategoryID             null.String `db:"category_id" json:"categoryID"`
	Name                   string      `db:"name" json:"name"`
	Pending                bool        `db:"pending" json:"pending"`
	HasReceipt             bool        `db:"has_receipt" json:"hasReceipt"`
	ReceiptType            null.String `db:"receipt_type" json:"receiptType"`
	PaymentChannel         string      `db:"payment_channel" json:"paymentChannel"` // ENUM: online, in store, other
	MerchantID             string      `db:"merchant_id" json:"merchantID"`
	MerchantName           null.String `json:"merchant_name"`
	UnofficialCurrencyCode null.String `db:"unofficial_currency_code" json:"unofficialCurrencyCode"`
	ISOCurrencyCode        null.String `db:"iso_currency_code" json:"isoCurrencyCode"`
	Amount                 float64     `db:"amount" json:"amount"`
	TransactionCode        null.String `db:"transaction_code" json:"transactionCode"` // ENUM (atm, bank charge, bill payment, cash, cashback, cheque, direct debit, interest, purchase, standing order, transfer, null)
	AuthorizedDate         null.Time   `db:"authorized_date" json:"authorizedDate"`
	AuthorizedDateTime     null.Time   `db:"authorized_datetime" json:"authorizedDateTime"`
	Date                   time.Time   `db:"date" json:"date"`
	DateTime               null.Time   `db:"datetime" json:"dateTime" diff:"-"`
	DeletedAt              null.Time   `db:"deleted_at" json:"deletedAt" diff:"-"`
	HiddenAt               null.Time   `db:"hidden_at" json:"hiddenAt" diff:"-"`
	CreatedAt              time.Time   `db:"created_at" json:"-" diff:"-"`
	UpdatedAt              time.Time   `db:"updated_at" json:"-" diff:"-"`

	Category    *PlaidCategory          `json:"category" diff:"-"`
	PaymentMeta *TransactionPaymentMeta `json:"transactionMeta" diff:"-"`
	Location    *TransactionLocation    `json:"location" diff:"-"`
}

func (r *Transaction) Filename() string {
	return fmt.Sprintf("%s.pdf", r.TransactionID)
}

type TransactionCategory struct {
	CategoryID string      `db:"category_id" json:"categoryID"`
	Category   SliceString `db:"category" json:"category"`
	Count      uint        `db:"count" json:"count"`
}

type TransactionReceipt struct {
	Get null.String
	Put null.String
}

type TransactionFilter struct {
	FromTransactionID null.String
	Limit             null.Uint64
	StartDate         null.Time
	EndDate           null.Time
	OnDate            null.Time
	DateInclusive     null.Bool
	AmountDir         null.Float64
	CategoryID        null.String
	MerchantID        null.String
}

func (f *TransactionFilter) BuildFromURLValues(values url.Values) error {
	categoryID := values.Get("categoryID")
	if categoryID != "" {
		f.CategoryID = null.NewString(categoryID, true)
	}

	fromTransactionID := values.Get("fromTransactionID")
	if fromTransactionID != "" {
		f.FromTransactionID = null.NewString(fromTransactionID, true)
	}

	limit := values.Get("limit")
	if limit != "" {
		parsedLimit, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			return err
		}

		f.Limit = null.Uint64From(parsedLimit)
	}

	startDate := values.Get("startDate")
	if startDate != "" {

		parsedDate, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return err
		}

		f.StartDate = null.NewTime(parsedDate, true)

	}

	endDate := values.Get("endDate")
	if endDate != "" {

		parsedDate, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return err
		}

		f.EndDate = null.NewTime(parsedDate, true)

	}

	dateInclusive := values.Get("dateInclusive")
	if dateInclusive != "" {

		parsedBool, err := strconv.ParseBool(dateInclusive)
		if err != nil {
			return err
		}

		f.DateInclusive = null.NewBool(parsedBool, true)

	}

	onDate := values.Get("onDate")
	if onDate != "" {
		parsedDate, err := time.Parse("2006-01-02", onDate)
		if err != nil {
			return err
		}

		f.OnDate = null.NewTime(parsedDate, true)
	}

	transactionType := values.Get("transactionType")
	if transactionType != "" {
		if transactionType == "expenses" {
			f.AmountDir = null.NewFloat64(-1, true)
		}
		if transactionType == "income" {
			f.AmountDir = null.NewFloat64(1, true)
		}
	}

	return nil
}

func (t *Transaction) FromPlaidTransaction(transaction plaid.Transaction) {

	t.AccountID = transaction.AccountID
	t.TransactionID = transaction.ID
	t.PendingTransactionID = null.NewString(transaction.PendingTransactionID, transaction.PendingTransactionID != "")
	t.CategoryID = null.NewString(transaction.CategoryID, transaction.CategoryID != "")
	t.Name = transaction.Name
	t.Pending = transaction.Pending
	t.PaymentChannel = transaction.PaymentChannel
	t.MerchantName = null.NewString(transaction.MerchantName, transaction.MerchantName != "")
	t.UnofficialCurrencyCode = null.NewString(transaction.UnofficialCurrencyCode, transaction.UnofficialCurrencyCode != "")
	t.ISOCurrencyCode = null.NewString(transaction.ISOCurrencyCode, transaction.ISOCurrencyCode != "")
	t.Amount = transaction.Amount
	authorizedDate, err := time.Parse("2006-01-02", transaction.AuthorizedDate)
	t.AuthorizedDate = null.NewTime(authorizedDate, err == nil)
	date, err := time.Parse("2006-01-02", transaction.Date)
	if err == nil {
		t.Date = date
	}

	t.Location = &TransactionLocation{

		Address:     null.NewString(transaction.Location.Address, transaction.Location.Address != ""),
		City:        null.NewString(transaction.Location.City, transaction.Location.City != ""),
		Lat:         transaction.Location.Lat,
		Lon:         transaction.Location.Lon,
		Region:      null.NewString(transaction.Location.Region, transaction.Location.Region != ""),
		StoreNumber: null.NewString(transaction.Location.StoreNumber, transaction.Location.StoreNumber != ""),
		PostalCode:  null.NewString(transaction.Location.PostalCode, transaction.Location.PostalCode != ""),
		Country:     null.NewString(transaction.Location.Country, transaction.Location.Country != ""),
	}

}

func (t *Transaction) FromUpdateTransactionInput(input *UpdateTransactionInput) {

	if input.Name.Valid {
		t.Name = input.Name.String
	}

	if input.MerchantID.Valid {
		t.MerchantID = input.MerchantID.String
	}

	t.CategoryID = input.CategoryID

}

type UpdateTransactionInput struct {
	Name       null.String
	MerchantID null.String
	CategoryID null.String
}

type Categories []string

func (s *Categories) Scan(value interface{}) error {

	switch data := value.(type) {
	case []byte:
		err := json.Unmarshal(data, s)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("failed to unmarshal categories, unsupport value type %T", value)
	}

	return nil

}

func (t Categories) Value() (driver.Value, error) {

	s := make([]string, 0, len(t))
	for _, tv := range t {
		s = append(s, tv)
	}

	if len(t) == 0 {
		return "", nil
	}

	return strings.Join(s, ","), nil

}

type TransactionLocation struct {
	ItemID        string      `db:"item_id" json:"-"`
	TransactionID string      `db:"transaction_id" json:"transactionID"`
	Address       null.String `db:"address" json:"address"`
	City          null.String `db:"city" json:"city"`
	Region        null.String `db:"region" json:"region"`
	PostalCode    null.String `db:"postal_code" json:"postalCode"`
	Country       null.String `db:"country" json:"country"`
	Lat           float64     `db:"lat" json:"lat"`
	Lon           float64     `db:"lon" json:"lon"`
	StoreNumber   null.String `db:"store_number" json:"storeNumber"`
	CreatedAt     time.Time   `db:"created_at" json:"-"`
	UpdatedAt     time.Time   `db:"updated_at" json:"-"`
}

func (tl *TransactionLocation) IsEmpty() bool {
	return tl.Address.Valid
}

type TransactionPaymentMeta struct {
	TransactionID    string      `db:"transaction_id" json:"transactionID"`
	ReferenceNumber  null.String `db:"reference_number" json:"referenceNumber"`
	PPDID            null.String `db:"ppd_id" json:"ppdID"`
	Payee            null.String `db:"payee" json:"payee"`
	ByOrderOf        null.String `db:"by_order_of" json:"byOrderOf"`
	Payer            null.String `db:"payer" json:"payer"`
	PaymentMethod    null.String `db:"payment_method" json:"paymentMethod"`
	PaymentProcessor null.String `db:"payment_processor" json:"paymentProcessor"`
	Reason           null.String `db:"reason" json:"reason"`
	CreatedAt        time.Time   `db:"created_at" json:"-"`
	UpdatedAt        time.Time   `db:"updated_at" json:"-"`
}
