package main

import (
	"context"
	//"encoding/json"
	//"time"
	"flag"
	"fmt"
	"os"
	"os/signal"
	//"strconv"
	"syscall"
	"bufio"

	//helpers "github.com/skolzkyi/antibruteforce/helpers"
	"github.com/skolzkyi/antibruteforce/internal/logger_cli"
	
)

type inputData struct {
	scanner *bufio.Scanner
}

func(id *inputData) Init() {
	id.scanner=bufio.NewScanner(os.Stdin)
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
	}
	//fmt.Println("config: ", config)
	log, err := logger_cli.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	inData:=inputData{}
	inData.Init()

	comContr:=CommandControllerNew()
	comContr.Init(config.GetAddress() + ":" + config.GetPort(),log)

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
			rawCommand:=inData.scanner.Text()
			if rawCommand == "" {
				continue
			}
			if rawCommand=="exit" {
				fmt.Println("bye")
				log.Info("antibruteforce-cli  down")
				os.Exit(1) //nolint:gocritic
				//cancel()
			}
			output:=comContr.processCommand(rawCommand)
			fmt.Println(output)
		}
	}
}