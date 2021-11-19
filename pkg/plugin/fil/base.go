package fil

import (
	"errors"

	"github.com/filecoin-project/go-state-types/crypto"
)

var ErrSignTypeInvalid = errors.New("sign type invalid")

func SignType(signType string) (crypto.SigType, error) {
	switch signType {
	case "secp256k1":
		return crypto.SigTypeBLS, nil
	case "bls":
		return crypto.SigTypeBLS, nil
	default:
		return crypto.SigTypeUnknown, ErrSignTypeInvalid
	}
}
