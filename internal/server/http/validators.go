package internalhttp

import (
	"errors"
	//"time"

	//storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

type RequestAuth struct {
	Login    string
	Password string
	IP       string
}

var (
	ErrVoidLogin          	= errors.New("login is void")
	ErrVoidPassword         = errors.New("password is void")
	ErrVoidIP       		= errors.New("IP is void")
)

func SimpleRequestValidator(login string, password string, IP string) (RequestAuth, error) { //nolint:lll
	request := RequestAuth{Login: login, Password: password, IP: IP} //nolint:lll
	switch {
	case request.Login == "":
		return RequestAuth{}, ErrVoidLogin 
	case request.Password == "":
		return RequestAuth{}, ErrVoidPassword 
	case request.IP == "":
		return RequestAuth{}, ErrVoidIP
	default:
	}

	return request, nil
}
