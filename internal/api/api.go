package api

import (
	_ "embed"
)

//go:embed openapi.json
var ApiSpec []byte
