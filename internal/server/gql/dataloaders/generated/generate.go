//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 AccountsByItemIDLoader string []*github.com/ddouglas/ledger.Account
//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 InstitutionLoader string *github.com/ddouglas/ledger.PlaidInstitution
//go:generate go run github.com/ddouglas/dataloaden@v0.4.0 CategoryLoader string *github.com/ddouglas/ledger.PlaidCategory

package generated
