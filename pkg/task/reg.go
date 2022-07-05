package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
)

const (
	// unit seconds
	getTransactionsTimeout = 60 * time.Second
	// unit seconds
	updateTransactionsTimeout = 5 * time.Second
)

type tworker struct {
	// unit second
	interval int
	// unit second
	// timeout int
	handle func(name string, interval int)
}

var (
	ErrTaskAlreadyRegister = errors.New("task already register")

	tworkers = make(map[string]tworker)
)

// interval unit second
// generate random number [3, 6)
func register(taskName string, interval int, handle func(name string, interval int)) error {
	if _, ok := tworkers[taskName]; ok {
		return ErrTaskAlreadyRegister
	}

	tworkers[taskName] = tworker{
		interval: interval,
		handle:   handle,
	}

	return nil
}

func fatalf(prefix, template string, args ...interface{}) {
	logger.Sugar().Fatalf(fmt.Sprintf("%s %v", prefix, template), args...)
}

func errorf(prefix, template string, args ...interface{}) {
	logger.Sugar().Errorf(fmt.Sprintf("%s %v", prefix, template), args...)
}

func warnf(prefix, template string, args ...interface{}) {
	logger.Sugar().Warnf(fmt.Sprintf("%s %v", prefix, template), args...)
}

func infof(prefix, template string, args ...interface{}) {
	logger.Sugar().Infof(fmt.Sprintf("%s %v", prefix, template), args...)
}

func Run() {
	for name, tf := range tworkers {
		logger.Sugar().Infof("run task: %v duration: %v", name, tf.interval)
		go tf.handle(name, tf.interval)
	}
}
