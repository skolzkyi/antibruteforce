package app

import (
	"errors"
	"strconv"
	"strings"
	//"time"
	"fmt"

	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

var (
	ErrVoidLogin    = errors.New("login is void")
	ErrVoidPassword = errors.New("password is void")
	ErrVoidIP       = errors.New("IP is void")

	ErrBadIP = errors.New("IP structure is bad")

	ErrVoidMask      = errors.New("mask is void")
	ErrIncorrectMask = errors.New("mask is incorrect")
)

func SimpleRequestValidator(login string, password string, IP string) (storageData.RequestAuth, error) { //nolint:lll
	request := storageData.RequestAuth{Login: login, Password: password, IP: IP} //nolint:lll
	err := checkIP(IP, 0, 255)
	switch {
	case err != nil:
		return storageData.RequestAuth{}, err
	case request.Login == "":
		return storageData.RequestAuth{}, ErrVoidLogin
	case request.Password == "":
		return storageData.RequestAuth{}, ErrVoidPassword
	default:
	}

	return request, nil
}

func SimpleIPDataValidator(IPData storageData.StorageIPData, isAllRequest bool) error { //nolint:lll
	var err error
	if !isAllRequest {
		err = checkIP(IPData.IP, 0, 255)
	}
	switch {
	case err != nil:
		return err
	case IPData.IP == "":
		return ErrVoidIP
	case IPData.Mask == 0 && !isAllRequest:
		return ErrVoidMask
	case IPData.Mask < 0 || IPData.Mask > 31:
		return ErrIncorrectMask
	default:
	}

	return nil
}

func checkIP(IP string, low int, high int) error {
	oktets := strings.Split(IP, ".")
	if len(oktets) != 4 {
		fmt.Println("IP: ", IP, " oktets: ", len(oktets))
		return ErrBadIP
	}
	for _, curOktet := range oktets {
		intOktet, err := strconv.Atoi(curOktet)
		if err != nil {
			return err
		}
		if intOktet < low || intOktet > high {
			fmt.Println("IP: ", IP, " bad oktet: ", curOktet)
			return ErrBadIP
		}
	}
	return nil
}
