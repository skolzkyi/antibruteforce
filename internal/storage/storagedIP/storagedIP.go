package storagedIP

import (
	"errors"
	"strconv"
	"time"
	"go.uber.org/zap"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
)

var (
	ErrNoRecord       = errors.New("record not searched")
	ErrStorageTimeout = errors.New("storage timeout")
)

type Config interface {
	Init(path string) error
	GetServerURL() string
	GetAddress() string
	GetPort() string
	GetServerShutdownTimeout() time.Duration
	GetDBName() string
	GetDBUser() string
	GetDBPassword() string
	GetDBConnMaxLifetime() time.Duration
	GetDBMaxOpenConns() int
	GetDBMaxIdleConns() int
	GetDBTimeOut() time.Duration
	GetDBAddress() string
	GetDBPort() string
	GetResisAddress() string 
	GetRedisPort() string 
	GetLimitFactorLogin() int 
	GetLimitFactorPassword() int 
	GetLimitFactorIP() int 
	GetLimitTimeCheck() time.Duration
}

type Logger interface {
	Info(msg string)
	Warning(msg string)
	Error(msg string)
	Fatal(msg string)
	GetZapLogger() *zap.SugaredLogger
}

type storagedIP struct {
	IP                    string
	ID                    int
}

func (ip *storagedIP) String() string {
	res := helpers.StringBuild("[ID: ", strconv.Itoa(ip.ID), ", IP: ", ip.IP, "]") 
	return res
}
