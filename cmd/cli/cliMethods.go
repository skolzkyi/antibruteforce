package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

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

type InputTag struct {
	Tag string
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
		return cc.addToWhiteList(commandData)
	case "removefromwhitelist", "rwl":
		return cc.removeFromWhiteList(commandData)
	case "isinwhitelist", "iwl":
		return cc.isInWhiteList(commandData)
	case "allinwhitelist", "allwl":
		return cc.allInWhiteList()
	case "addtoblacklist", "abl":
		return cc.addToBlackList(commandData)
	case "removefromblacklist", "rbl":
		return cc.removeFromBlackList(commandData)
	case "isinblacklist", "ibl":
		return cc.isInBlackList(commandData)
	case "allinblacklist", "allbl":
		return cc.allInBlackList()
	case "request", "req":
		return cc.request(commandData)
	case "clearbucketbylogin", "cbl":
		return cc.clearBucketByLogin(commandData)
	case "clearbucketbyip", "cbip":
		return cc.clearBucketByIP(commandData)
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

func (cc *CommandController) addToWhiteList(arg []string) string {
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

	url := helpers.StringBuild("http://", cc.address, "/whitelist/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
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

	mes := "subnet add to whitelist successful"
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) removeFromWhiteList(arg []string) string {
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

	url := helpers.StringBuild("http://", cc.address, "/whitelist/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := "subnet remove from whitelist successful"
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) isInWhiteList(arg []string) string {
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

	url := helpers.StringBuild("http://", cc.address, "/whitelist/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := "subnet in whitelist: " + answer.Message.Text
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) allInWhiteList() string {
	url := helpers.StringBuild("http://", cc.address, "/whitelist/")

	jsonStr := []byte(`{"IP":"ALL","Mask":0}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := "whitelist:\n" + result
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) addToBlackList(arg []string) string {
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

	url := helpers.StringBuild("http://", cc.address, "/blacklist/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
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

	mes := "subnet add to blacklist successful"
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) removeFromBlackList(arg []string) string {
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

	url := helpers.StringBuild("http://", cc.address, "/blacklist/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := "subnet remove from blacklist successful"
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) isInBlackList(arg []string) string {
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

	url := helpers.StringBuild("http://", cc.address, "/blacklist/")

	jsonStr := []byte(`{"IP":"` + subArgs[0] + `","Mask":` + subArgs[1] + `}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := "subnet in blacklist: " + answer.Message.Text
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) allInBlackList() string {
	url := helpers.StringBuild("http://", cc.address, "/blacklist/")

	jsonStr := []byte(`{"IP":"ALL","Mask":0}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := "blacklist:\n" + result
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) request(arg []string) string {
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
	resp, err := client.Do(req)
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

func (cc *CommandController) clearBucketByLogin(arg []string) string {
	if len(arg) != 2 {
		errStr := "error: " + ErrBadArgCount.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	url := helpers.StringBuild("http://", cc.address, "/clearbucketbylogin/")

	jsonStr := []byte(`{"Tag":"` + arg[1] + `"}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := `Login bucket "` + arg[1] + `" cleared`
	cc.logger.Info(mes)

	return mes
}

func (cc *CommandController) clearBucketByIP(arg []string) string {
	if len(arg) != 2 {
		errStr := "error: " + ErrBadArgCount.Error()
		cc.logger.Error(errStr)

		return errStr
	}

	url := helpers.StringBuild("http://", cc.address, "/clearbucketbyip/")

	jsonStr := []byte(`{"Tag":"` + arg[1] + `"}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errStr := "error: " + err.Error()
		cc.logger.Error(errStr)

		return errStr
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	mes := `IP bucket "` + arg[1] + `" cleared`
	cc.logger.Info(mes)

	return mes
}
