package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	"github.com/skolzkyi/antibruteforce/internal/loggercli"
)

var ErrABNotAvailable = errors.New("antibruteforce not available")

type inputData struct {
	scanner *bufio.Scanner
}

func (id *inputData) Init() {
	id.scanner = bufio.NewScanner(os.Stdin)
}

var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", "./configs/", "Path to config_cli.env")
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

	log, err := loggercli.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	err = pingAB(config.GetAddress() + ":" + config.GetPort())
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
	inData := inputData{}
	inData.Init()

	comContr := CommandControllerNew()
	comContr.Init(config.GetAddress()+":"+config.GetPort(), log)

	log.Info("antibruteforceAddr: " + config.GetAddress() + ":" + config.GetPort())
	log.Info("antibruteforce-cli up")
	fmt.Println("Welcome to antibrutforce-cli!")
	fmt.Println(`Use "help" command for an overview of available commands`)

	for {
		select {
		case <-ctx.Done():
			log.Info("antibruteforce-cli  down")
			os.Exit(1) //nolint:gocritic, nolintlint
		default:
			inData.scanner.Scan()
			rawCommand := inData.scanner.Text()
			if rawCommand == "" {
				continue
			}
			if rawCommand == "exit" {
				fmt.Println("bye")
				log.Info("antibruteforce-cli  down")
				os.Exit(1) //nolint:gocritic, nolintlint
			}
			output := comContr.processCommand(rawCommand)
			fmt.Println(output)
		}
	}
}

func pingAB(address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url:=helpers.StringBuild("http://", address, "/")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(respBody) != "test" {
		return ErrABNotAvailable
	}

	return nil
}
