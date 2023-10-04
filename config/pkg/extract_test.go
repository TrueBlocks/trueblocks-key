package extract

import (
	"fmt"
	"testing"
)

func TestEnvVariables(t *testing.T) {
	t.Setenv(fmt.Sprintf("%s_DATABASE_DEFAULT_HOST", prefix), "localhost")
	t.Setenv(fmt.Sprintf("%s_DATABASE_DEFAULT_PORT", prefix), "5324")

	config, err := Get("")
	if err != nil {
		t.Fatal(err)
	}

	if host := config.Database["default"].Host; host != "localhost" {
		t.Fatal("invalid host:", host)
	}
	if port := config.Database["default"].Port; port != 5324 {
		t.Fatal("invalid port:", port)
	}
}
