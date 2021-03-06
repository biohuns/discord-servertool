package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/xerrors"
	_ "net/http/pprof"
)

var exit = make(chan int)

func main() {
	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()

	log, err := initLogService()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	if err := listenStart(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Info("start listen")

	if err := batchStart(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Info("start batch")

	code := <-exit
	os.Exit(code)
}

func listenStart() error {
	message, err := initMessageService()
	if err != nil {
		return xerrors.Errorf("failed to init message service: %w", err)
	}

	if err := message.Start(); err != nil {
		return xerrors.Errorf("failed to start message service: %w", err)
	}

	return nil
}

func batchStart() error {
	batch, err := initBatchService()
	if err != nil {
		return xerrors.Errorf("failed to init batch service: %w", err)
	}

	go batch.Start()

	return nil
}
