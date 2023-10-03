package extract

import (
	"fmt"
	"testing"
)

func TestEnvVariables(t *testing.T) {
	t.Setenv(fmt.Sprintf("%s_DATABASE_HOST", prefix), "localhost")
	t.Setenv(fmt.Sprintf("%s_DATABASE_PORT", prefix), "5324")

	config, err := Get("")
	if err != nil {
		t.Fatal(err)
	}

	if host := config.Database.Host; host != "localhost" {
		t.Fatal("invalid host:", host)
	}
	if port := config.Database.Port; port != 5324 {
		t.Fatal("invalid port:", port)
	}
}
