package task

import "github.com/NpoolPlatform/message/npool/sphinxproxy"

func Abort(err error) sphinxproxy.TransactionState {
	if err == nil {
		return sphinxproxy.TransactionState_TransactionStateDone
	}

	return sphinxproxy.TransactionState_TransactionStateFail
}
