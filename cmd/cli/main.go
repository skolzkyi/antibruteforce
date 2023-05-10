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

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	"github.com/skolzkyi/antibruteforce/internal/logger_cli"
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

	log, err := logger_cli.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	err = pingAB(config.GetAddress() + ":" + config.GetPort())
	if err != nil {
		log.Fatal(err.Error())
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
			os.Exit(1) //nolint:gocritic
		default:
			inData.scanner.Scan()
			rawCommand := inData.scanner.Text()
			if rawCommand == "" {
				continue
			}
			if rawCommand == "exit" {
				fmt.Println("bye")
				log.Info("antibruteforce-cli  down")
				os.Exit(1) //nolint:gocritic
			}
			output := comContr.processCommand(rawCommand)
			fmt.Println(output)
		}
	}
}

func pingAB(address string) error {
	url := helpers.StringBuild("http://", address, "/")

	resp, err := http.Get(url)
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
