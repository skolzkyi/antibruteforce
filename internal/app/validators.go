package app

import (
	"errors"
	"strconv"
	"strings"

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

func SimpleRequestValidator(login string, password string, ip string) (storageData.RequestAuth, error) {
	request := storageData.RequestAuth{Login: login, Password: password, IP: ip}
	err := checkIP(ip, 0, 255)
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

func SimpleIPDataValidator(ipData storageData.StorageIPData, isAllRequest bool) error {
	var err error
	if !isAllRequest {
		err = checkIP(ipData.IP, 0, 255)
	}
	switch {
	case err != nil:
		return err
	case ipData.IP == "":
		return ErrVoidIP
	case ipData.Mask == 0 && !isAllRequest:
		return ErrVoidMask
	case ipData.Mask < 0 || ipData.Mask > 31:
		return ErrIncorrectMask
	default:
	}

	return nil
}

func checkIP(ip string, low int, high int) error {
	oktets := strings.Split(ip, ".")
	if len(oktets) != 4 {
		return ErrBadIP
	}
	for _, curOktet := range oktets {
		intOktet, err := strconv.Atoi(curOktet)
		if err != nil {
			return err
		}
		if intOktet < low || intOktet > high {
			return ErrBadIP
		}
	}

	return nil
}
