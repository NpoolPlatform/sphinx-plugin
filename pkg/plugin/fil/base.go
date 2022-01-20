package fil

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/filecoin-project/go-state-types/crypto"
)

func SignType(signType string) (crypto.SigType, error) {
	switch signType {
	case "secp256k1":
		return crypto.SigTypeSecp256k1, nil
	case "bls":
		return crypto.SigTypeBLS, nil
	default:
		return crypto.SigTypeUnknown, env.ErrSignTypeInvalid
	}
}
