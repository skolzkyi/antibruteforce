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
	//"time"

	_ "github.com/go-sql-driver/mysql"
	//nolint:gci,gofmt,gofumpt,nolintlint
	"github.com/skolzkyi/antibruteforce/internal/app"
	"github.com/skolzkyi/antibruteforce/internal/logger"
	
	internalhttp "github.com/skolzkyi/antibruteforce/internal/server/http"
	
	SQLstorage "github.com/skolzkyi/antibruteforce/internal/storage/sql"
)

var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", "./configs/", "Path to config.env")
}

func main() {
	//time.Sleep(30*time.Second)
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	err := config.Init(configFilePath)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config: ", config)
	log, err := logger.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
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

	//TODO init redis
	
	calendar := app.New(log, storage)

	server := internalhttp.NewServer(log, calendar, &config)

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
