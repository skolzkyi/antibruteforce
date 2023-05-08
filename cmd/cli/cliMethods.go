package main

import (
	//"context"
	"encoding/json"
	"net/http"
	"bytes"
	"io"
	//"flag"
	//"fmt"
	//"os"
	//"os/signal"
	"strconv"
	//"syscall"
	//"bufio"
	"errors"
	"strings"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
	//"github.com/skolzkyi/antibruteforce/internal/logger"
	
)

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
}

var(
	ErrUnSupCommand =errors.New("unsupported command")
	ErrBadArgCount 	=errors.New("bad argument count")
	ErrBadArgument	=errors.New("bad argument structure")
)

func CommandControllerNew() *CommandController{
	return &CommandController{}
}

func (cc *CommandController)Init(address string) {
	cc.address=address
}

func(cc *CommandController)processCommand(rawCommand string) string{
	commandData := strings.Split(rawCommand, " ")
	for i:=range commandData {
		commandData[i] = strings.ToLower(strings.TrimSpace(commandData[i]))
	}
	
	switch commandData[0] {
		case "help":
			return cc.help()
		case "addtowhitelist","awl":
			return cc.addToWhiteList(commandData)
		case "removefromwhitelist","rwl":
			return cc.removeFromWhiteList(commandData)
		case "isinwhitelist","iwl":
			return cc.isInWhiteList(commandData)
		case "allinwhitelist","allwl":
			return cc.allInWhiteList()
		case "addtoblacklist","abl":
			return cc.addToBlackList(commandData)
		case "removefromblacklist","rbl":
			return cc.removeFromBlackList(commandData)
		case "isinblacklist","ibl":
			return cc.isInBlackList(commandData)
		case "allinblacklist","allbl":
			return cc.allInBlackList()
		case "request","req":
			return cc.request(commandData)
		case "clearbucketbylogin","cbl":
			return cc.clearBucketByLogin(commandData)
		case "clearbucketbyip", "cbip":
			return cc.clearBucketByIP(commandData)
		default:
	}
	return "error: " + ErrUnSupCommand.Error()
}
func(cc *CommandController)help() string {
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

func(cc *CommandController) addToWhiteList(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	subArgs:=strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		return "error: " + ErrBadArgument.Error()
	}

	url := helpers.StringBuild("http://", cc.address, "/whitelist/")
	
	jsonStr := []byte(`{"IP":"`+subArgs[0]+`","Mask":`+subArgs[1]+`}`)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}	
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
		
	answer :=outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Text!="OK!"{
		return "error: " + answer.Text
	}

	return "subnet add to whitelist succesful"
}

func(cc *CommandController)removeFromWhiteList(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	subArgs:=strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		return "error: " + ErrBadArgument.Error()
	}

	url := helpers.StringBuild("http://", cc.address, "/whitelist/")
	
	jsonStr := []byte(`{"IP":"`+subArgs[0]+`","Mask":`+subArgs[1]+`}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
		
	answer :=outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Text!="OK!"{
		return "error: " + answer.Text
	}

	return "subnet remove from whitelist succesful"
}

func(cc *CommandController)isInWhiteList(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	subArgs:=strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		return "error: " + ErrBadArgument.Error()
	}

	url := helpers.StringBuild("http://", cc.address, "/whitelist/")
	
	jsonStr := []byte(`{"IP":"`+subArgs[0]+`","Mask":`+subArgs[1]+`}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
	//fmt.Println("iwl body: ", string(respBody))
	answer :=IPListAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Message.Code!=0{
		return "error: " + answer.Message.Text
	}

	return "subnet in whitelist: " + answer.Message.Text
}

func(cc *CommandController)allInWhiteList() string {
	url := helpers.StringBuild("http://", cc.address, "/whitelist/")
	
	jsonStr := []byte(`{"IP":"ALL","Mask":0}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
	//fmt.Println("iwl body: ", string(respBody))
	answer :=IPListAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Message.Code!=0{
		return "error: " + answer.Message.Text
	}
	result := ""
	for _,curIPSubNet:=range answer.IPList {
		result = helpers.StringBuild(result,curIPSubNet.IP,"/",strconv.Itoa(curIPSubNet.Mask),"\n")
	}

	return "whitelist:\n" + result
}

func(cc *CommandController) addToBlackList(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	subArgs:=strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		return "error: " + ErrBadArgument.Error()
	}

	url := helpers.StringBuild("http://", cc.address, "/blacklist/")
	
	jsonStr := []byte(`{"IP":"`+subArgs[0]+`","Mask":`+subArgs[1]+`}`)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}	
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
		
	answer :=outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Text!="OK!"{
		return "error: " + answer.Text
	}

	return "subnet add to blacklist succesful"
}

func(cc *CommandController)removeFromBlackList(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	subArgs:=strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		return "error: " + ErrBadArgument.Error()
	}

	url := helpers.StringBuild("http://", cc.address, "/blacklist/")
	
	jsonStr := []byte(`{"IP":"`+subArgs[0]+`","Mask":`+subArgs[1]+`}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
		
	answer :=outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Text!="OK!"{
		return "error: " + answer.Text
	}

	return "subnet remove from blacklist succesful"
}

func(cc *CommandController)isInBlackList(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	subArgs:=strings.Split(arg[1], "/")

	if len(subArgs) != 2 {
		return "error: " + ErrBadArgument.Error()
	}

	url := helpers.StringBuild("http://", cc.address, "/blacklist/")
	
	jsonStr := []byte(`{"IP":"`+subArgs[0]+`","Mask":`+subArgs[1]+`}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
	//fmt.Println("iwl body: ", string(respBody))
	answer :=IPListAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Message.Code!=0{
		return "error: " + answer.Message.Text
	}

	return "subnet in blacklist: " + answer.Message.Text
}

func(cc *CommandController)allInBlackList() string {
	url := helpers.StringBuild("http://", cc.address, "/blacklist/")
	
	jsonStr := []byte(`{"IP":"ALL","Mask":0}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
	//fmt.Println("iwl body: ", string(respBody))
	answer :=IPListAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Message.Code!=0{
		return "error: " + answer.Message.Text
	}
	result := ""
	for _,curIPSubNet:=range answer.IPList {
		result = helpers.StringBuild(result,curIPSubNet.IP,"/",strconv.Itoa(curIPSubNet.Mask),"\n")
	}

	return "blacklist:\n" + result
}

func(cc *CommandController)request(arg []string) string {
	if len(arg) != 4 {
		return "error: " + ErrBadArgCount.Error()
	}
	
	url := helpers.StringBuild("http://", cc.address, "/request/")
	
	jsonStr := []byte(`{
		"Login":"`+arg[1]+`",
		"Password":"`+arg[2]+`",
		"IP":"`+arg[3]+`"
	}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
	//fmt.Println("iwl body: ", string(respBody))
	answer :=AuthorizationRequestAnswer{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
	var txtAnswer string
	if answer.Ok {
		txtAnswer="NO"
	}else{
		txtAnswer="YES"
	}

	return "is bruteforce: " + txtAnswer + "; notification: " + answer.Message
}
// req u0 p1 10.0.0.0
func(cc *CommandController)clearBucketByLogin(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	
	url := helpers.StringBuild("http://", cc.address, "/clearbucketbylogin/")
	
	jsonStr := []byte(`{"Tag":"`+arg[1]+`"}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
		
	answer :=outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Text!="OK!"{
		return "error: " + answer.Text
	}

	return `Login bucket "`+arg[1]+`" cleared`
}

func(cc *CommandController)clearBucketByIP(arg []string) string {
	if len(arg) != 2 {
		return "error: " + ErrBadArgCount.Error()
	}
	
	url := helpers.StringBuild("http://", cc.address, "/clearbucketbyip/")
	
	jsonStr := []byte(`{"Tag":"`+arg[1]+`"}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err!=nil {
		return "error: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err!=nil {
		return "error: " + err.Error()
	}
		
	answer :=outputJSON{}
	err = json.Unmarshal(respBody, &answer)
	if err!=nil {
		return "error: " + err.Error()
	}
    if answer.Text!="OK!"{
		return "error: " + answer.Text
	}

	return `IP bucket "`+arg[1]+`" cleared`
}