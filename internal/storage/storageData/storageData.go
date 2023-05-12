package storagedata

import (
	"errors"
	"strconv"
	"time"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	"go.uber.org/zap"
)

const WhiteListName string = "whitelist"
const BlackListName string = "blacklist"
var (
	ErrNoRecord       = errors.New("record not searched")
	ErrStorageTimeout = errors.New("storage timeout")
	ErrErrorBadListType = errors.New("bad list type")
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
	GetRedisAddress() string
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

type StorageIPData struct {
	IP   string
	Mask int
	ID   int
}

func (ip *StorageIPData) String() string {
	res := helpers.StringBuild("[ID: ", strconv.Itoa(ip.ID), ", IP: ", ip.IP, "]")

	return res
}

type RequestAuth struct {
	Login    string
	Password string
	IP       string
}

func (r *RequestAuth) String() string {
	res := helpers.StringBuild("[Login: ", r.Login, " Password: ", r.Password, ", IP: ", r.IP, "]")

	return res
}
