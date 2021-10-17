package main

import (
	"io/fs"
	"os"

	"github.com/ddouglas/ledger"
	"github.com/pkg/errors"
)

func getFrontendAssets() fs.FS {
	switch cfg.Env {
	case "production":
		return getProductionAssets()
	default:
		return getDevelopmentAssets()
	}
}

func getDevelopmentAssets() fs.FS {
	return os.DirFS("frontend")
}

func getProductionAssets() fs.FS {
	f, err := fs.Sub(ledger.Frontend, "frontend")
	if err != nil {
		panic(errors.Wrap(err, "failed to load frontend assets").Error())
	}

	return f
}
