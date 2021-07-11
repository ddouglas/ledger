package ledger

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/volatiletech/null"
)

type TransactionRepository interface {
}

type Transaction struct {
	ItemID                 string      `db:"item_id" json:"itemID"`
	AccountID              string      `db:"account_id" json:"accountID"`
	TransactionID          string      `db:"transaction_id" json:"transactionID"`
	PendingTransactionID   null.String `db:"pending_transaction_id" json:"pendingTransactionID"`
	CategoryID             null.String `db:"category_id" json:"categoryID"`
	Name                   string      `db:"name" json:"name"`
	Pending                bool        `db:"pending" json:"pending"`
	PaymentChannel         string      `db:"payment_channel" json:"paymentChannel"` // ENUM: online, in store, other
	MerchantName           null.String `db:"merchant_name" json:"merchantName"`
	Categories             Categories  `db:"categories" json:"categories"` // Array, needs to be converted to comma-delimited string going into DB and Slice comming out
	UnofficialCurrencyCode null.String `db:"unofficial_currency_code" json:"unofficialCurrencyCode"`
	ISOCurrencyCode        null.String `db:"iso_currency_code" json:"isoCurrencyCode"`
	Amount                 float64     `db:"amount" json:"amount"`
	TransactionCode        null.String `db:"transaction_code" json:"transactionCode"` // ENUM (atm, bank charge, bill payment, cash, cashback, cheque, direct debit, interest, purchase, standing order, transfer, null)
	AuthorizedDate         null.Time   `db:"authorized_date" json:"authorizedDate"`
	AuthorizedDateTime     null.Time   `db:"authorized_datetime" json:"authorizedDateTime"`
	Date                   time.Time   `db:"date" json:"date"`
	DateTime               null.Time   `db:"datetime" json:"dateTime"`
	CreatedAt              time.Time   `db:"created_at" json:"-"`
	UpdatedAt              time.Time   `db:"updated_at" json:"-"`

	PaymentMeta TransactionPaymentMeta `json:"transactionMeta"`
	Location    TransactionLocation    `json:"location"`
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
	TransactionID string       `db:"transaction_id" json:"transactionID"`
	Address       null.String  `db:"address" json:"address"`
	City          null.String  `db:"city" json:"city"`
	Region        null.String  `db:"region" json:"region"`
	PostalCode    null.String  `db:"postal_code" json:"postalCode"`
	Country       null.String  `db:"country" json:"country"`
	Lat           null.Float64 `db:"lat" json:"lat"`
	Lon           null.Float64 `db:"lon" json:"lon"`
	StoreNumber   null.String  `db:"store_number" json:"storeNumber"`
	CreatedAt     time.Time    `db:"created_at" json:"-"`
	UpdatedAt     time.Time    `db:"updated_at" json:"-"`
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
