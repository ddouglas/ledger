//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 AccountsByItemIDLoader string []*github.com/ddouglas/ledger.Account
//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 InstitutionLoader string *github.com/ddouglas/ledger.PlaidInstitution
//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 CategoryLoader string *github.com/ddouglas/ledger.PlaidCategory
//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 MerchantLoader string *github.com/ddouglas/ledger.Merchant
//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 MerchantAliasLoader string []*github.com/ddouglas/ledger.MerchantAlias

package generated
