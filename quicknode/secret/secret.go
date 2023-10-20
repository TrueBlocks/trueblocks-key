package secret

import (
	"encoding/json"
	"errors"

	awshelper "trueblocks.io/awshelper/pkg"
)

type authSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func FetchAuthSecret(secretId string) (secret *authSecret, err error) {
	if secretId == "" {
		err = errors.New("secretId is empty")
		return
	}

	encoded, err := awshelper.FetchSecret(secretId)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(encoded), secret); err != nil {
		return nil, err
	}

	return secret, nil
}
