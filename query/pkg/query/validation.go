package query

import "encoding/hex"

func validateAddress(address string) error {
	if len(address) != 42 {
		return ErrAddressIncorrect
	}
	if address[:2] != "0x" {
		return ErrAddressIncorrect
	}
	if _, err := hex.DecodeString(address[2:]); err != nil {
		return ErrAddressIncorrect
	}
	return nil
}
