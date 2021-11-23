package fil

import (
	"github.com/filecoin-project/go-state-types/crypto"
)

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
