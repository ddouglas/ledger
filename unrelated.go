package ledger

type AutoGenerated struct {
	Accounts []struct {
		AccountID string `json:"account_id"`
		Balances  struct {
			Available              int         `json:"available"`
			Current                int         `json:"current"`
			IsoCurrencyCode        string      `json:"iso_currency_code"`
			Limit                  interface{} `json:"limit"`
			UnofficialCurrencyCode interface{} `json:"unofficial_currency_code"`
		} `json:"balances"`
		Mask         string `json:"mask"`
		Name         string `json:"name"`
		OfficialName string `json:"official_name"`
		Subtype      string `json:"subtype"`
		Type         string `json:"type"`
	} `json:"accounts"`
	Transactions []struct {
		AccountID              string      `json:"account_id"`
		Amount                 float64     `json:"amount"`
		IsoCurrencyCode        string      `json:"iso_currency_code"`
		UnofficialCurrencyCode interface{} `json:"unofficial_currency_code"`
		Category               []string    `json:"category"`
		CategoryID             string      `json:"category_id"`
		Date                   string      `json:"date"`
		Datetime               interface{} `json:"datetime"`
		AuthorizedDate         string      `json:"authorized_date"`
		AuthorizedDatetime     interface{} `json:"authorized_datetime"`
		Location               struct {
			Address     string  `json:"address"`
			City        string  `json:"city"`
			Region      string  `json:"region"`
			PostalCode  string  `json:"postal_code"`
			Country     string  `json:"country"`
			Lat         float64 `json:"lat"`
			Lon         float64 `json:"lon"`
			StoreNumber string  `json:"store_number"`
		} `json:"location"`
		Name         string `json:"name"`
		MerchantName string `json:"merchant_name"`
		PaymentMeta  struct {
			ByOrderOf        interface{} `json:"by_order_of"`
			Payee            interface{} `json:"payee"`
			Payer            interface{} `json:"payer"`
			PaymentMethod    interface{} `json:"payment_method"`
			PaymentProcessor interface{} `json:"payment_processor"`
			PpdID            interface{} `json:"ppd_id"`
			Reason           interface{} `json:"reason"`
			ReferenceNumber  interface{} `json:"reference_number"`
		} `json:"payment_meta"`
		PaymentChannel       string      `json:"payment_channel"`
		Pending              bool        `json:"pending"`
		PendingTransactionID interface{} `json:"pending_transaction_id"`
		AccountOwner         interface{} `json:"account_owner"`
		TransactionID        string      `json:"transaction_id"`
		TransactionCode      interface{} `json:"transaction_code"`
		TransactionType      string      `json:"transaction_type"`
	} `json:"transactions"`
	Item struct {
		AvailableProducts     []string    `json:"available_products"`
		BilledProducts        []string    `json:"billed_products"`
		ConsentExpirationTime interface{} `json:"consent_expiration_time"`
		Error                 interface{} `json:"error"`
		InstitutionID         string      `json:"institution_id"`
		ItemID                string      `json:"item_id"`
		UpdateType            string      `json:"update_type"`
		Webhook               string      `json:"webhook"`
	} `json:"item"`
	TotalTransactions int    `json:"total_transactions"`
	RequestID         string `json:"request_id"`
}
