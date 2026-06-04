package request

import (
	"encoding/json"
	"errors"
)

type ZipcodePayload struct {
	CEP json.RawMessage `json:"cep"`
}

func ParseZipcode(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", errors.New("invalid zipcode")
	}

	var cep string
	if err := json.Unmarshal(raw, &cep); err != nil {
		return "", errors.New("invalid zipcode")
	}

	return cep, nil
}
