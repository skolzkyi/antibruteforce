package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	loggercli "github.com/skolzkyi/antibruteforce/internal/loggercli"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

const correctAnswerText string = "OK!"

type AuthorizationRequestAnswer struct {
	Message string
	Ok      bool
}

type outputJSON struct {
	Text string
	Code int
}

type IPListAnswer struct {
	IPList  []storageData.StorageIPData
	Message outputJSON
}

type CommandController struct {
	address string
	logger  *loggercli.LogWrap
}

var (
	ErrUnSupCommand = errors.New("unsupported command")
	ErrBadArgCount  = errors.New("bad argument count")
	ErrBadArgument  = errors.New("bad argument structure")
)

func CommandControllerNew() *CommandController {
	return &CommandController{}
}

func (cc *CommandController) Init(address string, logger *loggercli.LogWrap) {
	cc.address = address
	cc.logger = logger
}

func (cc *CommandController) processCommand(rawCommand string) string {
	cc.logger.Info("command: " + rawCommand)
	commandData := strings.Split(rawCommand, " ")
	for i := range commandData {
		commandData[i] = strings.ToLower(strings.TrimSpace(commandData[i]))
	}

	switch commandData[0] {
	case "help":
		return cc.help()
	case "addtowhitelist", "awl":
		return cc.addToList(commandData, "whitelist")
	case "removefromwhitelist", "rwl":
		return cc.removeFromList(commandData, "whitelist")
	case "isinwhitelist", "iwl":
		return cc.isInList(commandData, "whitelist")
	case "allinwhitelist", "allwl":
		return cc.allInList("whitelist")
	case "addtoblacklist", "abl":
		return cc.addToList(commandData, "blacklist")
	case "removefromblacklist", "rbl":
		return cc.removeFromList(commandData, "blacklist")
	case "isinblacklist", "ibl":
		return cc.isInList(commandData, "blacklist")
	case "allinblacklist", "allbl":
		return cc.allInList("blacklist")
	case "request", "req":
		return cc.request(commandData)
	case "clearbucketbylogin", "cbl":
		return cc.clearBucketByTag(commandData, "login")
	case "clearbucketbyip", "cbip":
		return cc.clearBucketByTag(commandData, "ip")
	default:
	}
	mes := "error: " + ErrUnSupCommand.Error()
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) help() string {
	return `
help - overview of available commands
exit - exit from CLI
long: AddToWhiteList [subnet], short: awl [subnet] - add subnet to whitelist
long: RemoveFromWhiteList [subnet], short: rwl [subnet] - remove subnet from whitelist
long: IsInWhiteList [subnet], short: iwl [subnet]- check subnet in whitelist
long: AllInWhiteList, short: allwl - print whitelist
long: AddToBlackList [subnet], short: abl [subnet] - add subnet to blacklist
long: RemoveFromBlackList [subnet], short: rbl [subnet] - remove subnet from blacklist
long: IsInBlackList [subnet], short: ibl [subnet] - check subnet in blacklist
long: AllInBlackList [subnet], short: allbl [subnet] - print whitelist
long: Request [login] [password] [IP], short: req [login] [password] [IP] - sends a request whether it is bruteforce
long: ClearBucketByLogin [login], short: cbl [login] - cleared login bucket in bucket storage
long: ClearBucketByIP [IP], short: cbip [IP] - cleared IP bucket in bucket storage`
}

func (cc *CommandController) addToList(arg []string, listname string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(arg) != 2 {
		errStr := "error: " + ErrBadArgCount.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	subArgs := strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		errStr := "error: " + ErrBadArgument.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	url := helpers.StringBuild("http://", cc.address, "/", listname, "/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	answer := outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	if answer.Text != correctAnswerText {
		errStr := "error: " + answer.Text
		cc.logger.Error(errStr)

		return errStr
	}

	mes := "subnet add to " + listname + " successful"
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) removeFromList(arg []string, listname string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(arg) != 2 {
		errStr := "error: " + ErrBadArgCount.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	subArgs := strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		errStr := "error: " + ErrBadArgument.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	url := helpers.StringBuild("http://", cc.address, "/"+listname+"/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	answer := outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	if answer.Text != correctAnswerText {
		errStr := "error: " + answer.Text
		cc.logger.Error(errStr)

		return errStr
	}

	mes := "subnet remove from " + listname + "successful"
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) isInList(arg []string, listname string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(arg) != 2 {
		errStr := "error: " + ErrBadArgCount.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	subArgs := strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		errStr := "error: " + ErrBadArgument.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	url := helpers.StringBuild("http://", cc.address, "/"+listname+"/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	answer := IPListAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	if answer.Message.Code != 0 {
		errStr := "error: " + answer.Message.Text
		cc.logger.Error(errStr)

		return errStr
	}

	mes := "subnet in " + listname + ": " + answer.Message.Text
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) allInList(listname string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := helpers.StringBuild("http://", cc.address, "/"+listname+"/")

	jsonStr := []byte(`{"IP":"ALL","Mask":0}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	answer := IPListAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	if answer.Message.Code != 0 {
		errStr := "error: " + answer.Message.Text
		cc.logger.Error(errStr)

		return errStr
	}
	result := ""
	for _, curIPSubNet := range answer.IPList {
		result = helpers.StringBuild(result, curIPSubNet.IP, "/", strconv.Itoa(curIPSubNet.Mask), "\n")
	}

	mes := listname + ":\n" + result
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) request(arg []string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(arg) != 4 {
		errStr := "error: " + ErrBadArgCount.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	url := helpers.StringBuild("http://", cc.address, "/request/")

	jsonStr := []byte(`{
		"Login":"` + arg[1] + `",
		"Password":"` + arg[2] + `",
		"IP":"` + arg[3] + `"
	}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	answer := AuthorizationRequestAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	var txtAnswer string
	if answer.Ok {
		txtAnswer = "NO"
	} else {
		txtAnswer = "YES"
	}

	mes := "is bruteforce: " + txtAnswer + "; notification: " + answer.Message
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) clearBucketByTag(arg []string, typeClear string) string {
	var urlSegByType string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(arg) != 2 {
		errStr := "error: " + ErrBadArgCount.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	switch typeClear {
	case "login":
		urlSegByType = "/clearbucketbylogin/"
	case "ip":
		urlSegByType = "/clearbucketbyip/"
	default:
		errStr := "error: " + ErrBadArgument.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	url := helpers.StringBuild("http://", cc.address, urlSegByType)

	jsonStr := []byte(`{"Tag":"` + arg[1] + `"}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	answer := outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	if answer.Text != correctAnswerText {
		errStr := "error: " + answer.Text
		cc.logger.Error(errStr)

		return errStr
	}

	mes := typeClear + ` bucket "` + arg[1] + `" cleared`
	cc.logger.Info(mes)

	return mes
}
