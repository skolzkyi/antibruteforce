package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	//nolint:gci,gofmt,gofumpt,nolintlint
	"github.com/skolzkyi/antibruteforce/internal/app"
	"github.com/skolzkyi/antibruteforce/internal/logger"

	internalhttp "github.com/skolzkyi/antibruteforce/internal/server/http"

	RedisStorage "github.com/skolzkyi/antibruteforce/internal/storage/redis"
	SQLstorage "github.com/skolzkyi/antibruteforce/internal/storage/sql"
)

var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", "./configs/", "Path to config.env")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()

		return
	}

	config := NewConfig()
	err := config.Init(configFilePath)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	log, err := logger.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Info("servAddr: " + config.GetAddress())
	var storage app.Storage
	ctxStor, cancelStore := context.WithTimeout(context.Background(), config.GetDBTimeOut())

	storage = SQLstorage.New()
	err = storage.Init(ctxStor, log, &config)
	if err != nil {
		cancelStore()
		log.Fatal("fatal error of inintialization SQL storage: " + err.Error())
	}
	redis := RedisStorage.New()
	err = redis.Init(ctxStor, log, &config)
	if err != nil {
		cancelStore()
		log.Fatal("fatal error of inintialization Redis storage: " + err.Error())
	}

	antibruteforce := app.New(log, storage, redis, &config)

	server := internalhttp.NewServer(log, antibruteforce, &config)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), config.GetServerShutdownTimeout())
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			log.Fatal("failed to stop http server: " + err.Error())
		}
	}()

	if err := server.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
