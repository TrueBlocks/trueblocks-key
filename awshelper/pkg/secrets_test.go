package awshelper

import (
	"fmt"
	"testing"
)

func TestUsernamePasswordSecret_FromString(t *testing.T) {
	secret := &UsernamePasswordSecret{}
	username := "testUser"
	password := "v)_Hc)o1*aa234"

	err := secret.FromString(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password))
	if err != nil {
		t.Fatal(err)
	}

	if u := secret.Username; u != username {
		t.Fatal("wrong username:", u)
	}
	if p := secret.Password; p != password {
		t.Fatal("wrong password:", p)
	}
}
