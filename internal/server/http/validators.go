package app

import (
	"errors"
	"time"

	storagedIP "github.com/skolzkyi/hwOTUS_YIA/hw12_13_14_15_calendar/internal/storage/storagedIP"
)

type Request struct {
	Login    string
	Password string
	IP       string
}

var (
	ErrVoidLogin          	= errors.New("login is void")
	ErrVoidPassword         = errors.New("password is void")
	ErrVoidIP       		= errors.New("IP is void")
)

func SimpleRequestValidator(login string, password string, IP string) (storage.Event, error) { //nolint:lll
	request := Request{Login: login, Password: password, IP: IP} //nolint:lll
	switch {
	case request.Login == "":
		return Request{}, ErrVoidLogin 
	case request.Password == "":
		return Request{}, ErrVoidPassword 
	case request.IP == "":
		return Request{}, ErrVoidIP
	default:
	}

	return request, nil
}
