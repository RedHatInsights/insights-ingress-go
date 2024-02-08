package api

import (
	_ "embed"
	"github.com/gofiber/template/django/v3"
)

func main() {
	// Create a new engine
	if (false) {
		django.New("./views", ".django")
	}
}

//go:embed openapi.json
var ApiSpec []byte
