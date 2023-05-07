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
	//"strconv"
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

type output interface{} 

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

func(cc *CommandController)processCommand(rawCommand string) output {
	commandData := strings.Split(rawCommand, " ")
	for i:=range commandData {
		commandData[i] = strings.ToLower(strings.TrimSpace(commandData[i]))
	}
	
	switch commandData[0] {
		case "help":

		case "addtowhitelist","awl":
			return cc.addToWhiteList(commandData)
		case "removefromwhitelist","rwl":
			return cc.removeFromWhiteList(commandData)
		case "isinwhitelist","iwl":
			return cc.isInWhiteList(commandData)
		case "allinwhitelist","allwl":

		case "addtoblacklist","abl":

		case "removefromblacklist","rbl":

		case "isinblacklist","ibl":

		case "allinblacklist","allbl":

		case "request","req":

		case "clearbucketbylogin","cbl":

		case "clearbucketbuip", "cbip":

		default:
	}
	return "error: " + ErrUnSupCommand.Error()
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