package app

import (
	"errors"
	//"time"

	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

var (
	ErrVoidLogin          	= errors.New("login is void")
	ErrVoidPassword         = errors.New("password is void")
	ErrVoidIP       		= errors.New("IP is void")

	ErrVoidMask      		= errors.New("mask is void")
)

func SimpleRequestValidator(login string, password string, IP string) (storageData.RequestAuth, error) { //nolint:lll
	request := storageData.RequestAuth{Login: login, Password: password, IP: IP} //nolint:lll
	switch {
	case request.Login == "":
		return storageData.RequestAuth{}, ErrVoidLogin 
	case request.Password == "":
		return storageData.RequestAuth{}, ErrVoidPassword 
	case request.IP == "":
		return storageData.RequestAuth{}, ErrVoidIP
	default:
	}

	return request, nil
}


func SimpleIPDataValidator(IPData storageData.StorageIPData, isAllRequest bool)  error { //nolint:lll
	switch {
	case IPData.IP == "":
		return  ErrVoidIP
	case IPData.Mask == 0 && IPData.IP != "ALL" && isAllRequest:
		return ErrVoidMask 
	default:
	}

	return nil
}