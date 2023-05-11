package app

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"go.uber.org/zap"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

var (
	ErrIPDataExistInBL = errors.New("IPData already exists in blacklist")
	ErrIPDataExistInWL = errors.New("IPData already exists in whitelist")
)

type App struct {
	logger              Logger
	storage             Storage
	bucketStorage       BStorage
	ticker              *time.Ticker
	periodic            time.Duration
	limitFactorLogin    int
	limitFactorPassword int
	limitFactorIP       int
}

type Logger interface {
	Info(msg string)
	Warning(msg string)
	Error(msg string)
	Fatal(msg string)
	GetZapLogger() *zap.SugaredLogger
}

type Storage interface {
	Init(ctx context.Context, logger storageData.Logger, config storageData.Config) error
	Close(ctx context.Context, logger storageData.Logger) error
	AddIPToWhiteList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) (int, error)
	RemoveIPInWhiteList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) error
	IsIPInWhiteList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) (bool, error)
	GetAllIPInWhiteList(ctx context.Context, logger storageData.Logger) ([]storageData.StorageIPData, error)
	AddIPToBlackList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) (int, error)
	RemoveIPInBlackList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) error
	IsIPInBlackList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) (bool, error)
	GetAllIPInBlackList(ctx context.Context, logger storageData.Logger) ([]storageData.StorageIPData, error)
}

type BStorage interface {
	Init(ctx context.Context, logger storageData.Logger, config storageData.Config) error
	SetBucketValue(ctx context.Context, logger storageData.Logger, key string, value int) error
	Close(ctx context.Context, logger storageData.Logger) error
	IncrementAndGetBucketValue(ctx context.Context, logger storageData.Logger, key string) (int64, error)
	FlushStorage(ctx context.Context, logger storageData.Logger) error
}

func New(logger Logger, storage Storage, bStorage BStorage, config storageData.Config) *App {
	app := App{
		logger:              logger,
		storage:             storage,
		bucketStorage:       bStorage,
		limitFactorLogin:    config.GetLimitFactorLogin(),
		limitFactorPassword: config.GetLimitFactorPassword(),
		limitFactorIP:       config.GetLimitFactorIP(),
		periodic:            config.GetLimitTimeCheck(),
	}

	return &app
}

func (a *App) InitBStorageAndLimits(ctx context.Context, config storageData.Config) error {
	return a.bucketStorage.Init(ctx, a.logger, config)
}

func (a *App) CloseBStorage(ctx context.Context) error {
	return a.bucketStorage.Close(ctx, a.logger)
}

func (a *App) CheckInputRequest(ctx context.Context, req storageData.RequestAuth) (bool, string, error) {
	ok, err := a.isIPInBlackListCheck(ctx, req.IP)
	if err != nil {
		message := helpers.StringBuild("CheckInputRequest isIPInBlackListCheck error: ", err.Error())
		a.logger.Error(message)

		return false, "", err
	}
	if ok {
		return false, "IP in blacklist", nil
	}
	ok, err = a.isIPInWhiteListCheck(ctx, req.IP)
	if err != nil {
		message := helpers.StringBuild("CheckInputRequest isIPInWhiteListCheck error: ", err.Error())
		a.logger.Error(message)

		return false, "", err
	}
	if ok {
		return true, "IP in whitelist", nil
	}
	countLogin, err := a.bucketStorage.IncrementAndGetBucketValue(ctx, a.logger, "l_"+req.Login)
	if err != nil {
		errBaseText := "CheckInputRequest IncrementAndGetBucketValue - Login error: "
		message := helpers.StringBuild(errBaseText, err.Error(), ", key: ", "l_"+req.Login)
		a.logger.Error(message)

		return false, "", err
	}

	if countLogin > int64(a.limitFactorLogin) {
		return false, "rate limit by login", nil
	}
	countPassword, err := a.bucketStorage.IncrementAndGetBucketValue(ctx, a.logger, "p_"+req.Password)
	if err != nil {
		errBaseText := "CheckInputRequest IncrementAndGetBucketValue - Password error: "
		message := helpers.StringBuild(errBaseText, err.Error(), ", key: ", "p_"+req.Password)
		a.logger.Error(message)

		return false, "", err
	}
	if countPassword > int64(a.limitFactorPassword) {
		return false, "rate limit by password", nil
	}

	countIP, err := a.bucketStorage.IncrementAndGetBucketValue(ctx, a.logger, "ip_"+req.IP)
	if err != nil {
		message := helpers.StringBuild("CheckInputRequest IncrementAndGetBucketValue - IP error: ", err.Error(), ", key: ", "ip_"+req.IP)
		a.logger.Error(message)

		return false, "", err
	}
	if countIP > int64(a.limitFactorIP) {
		return false, "rate limit by IP", nil
	}

	return true, "clear check", nil
}

func (a *App) RLTicker(ctx context.Context) {
	a.logger.Info("ticker start")
	a.ticker = time.NewTicker(a.periodic)
	go func() {
		for {
			select {
			case <-ctx.Done():
				a.logger.Info("ticker stop")

				break
			case <-a.ticker.C:
				a.bucketStorage.FlushStorage(ctx, a.logger)
				a.logger.Info("buckets flush")
			}
		}
	}()
}

func (a *App) ClearBucketByLogin(ctx context.Context, login string) error {
	err := a.bucketStorage.SetBucketValue(ctx, a.logger, "l_"+login, 0)
	if err != nil {
		message := helpers.StringBuild("ClearBucketByLogin error", err.Error(), " Login: ", login)
		a.logger.Error(message)

		return err
	}
	a.logger.Info(" login bucket clear, login: " + login)

	return nil
}

func (a *App) ClearBucketByIP(ctx context.Context, ip string) error {
	err := a.bucketStorage.SetBucketValue(ctx, a.logger, "ip_"+ip, 0)
	if err != nil {
		message := helpers.StringBuild("ClearBucketByIP error", err.Error(), " IP: ", ip)
		a.logger.Error(message)

		return err
	}
	a.logger.Info(" IP bucket clear, IP: " + ip)

	return nil
}

func (a *App) InitStorage(ctx context.Context, config storageData.Config) error {
	return a.storage.Init(ctx, a.logger, config)
}

func (a *App) CloseStorage(ctx context.Context) error {
	return a.storage.Close(ctx, a.logger)
}

// WHITELIST

func (a *App) AddIPToWhiteList(ctx context.Context, ipData storageData.StorageIPData) (int, error) {
	err := SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("AddIPToWhiteList validate IPData error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	ok, err := a.storage.IsIPInBlackList(ctx, a.logger, ipData)
	if err != nil {
		message := helpers.StringBuild("AddIPToWhiteList validate in blacklist IPData error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	if ok {
		return 0, ErrIPDataExistInBL
	}
	id, err := a.storage.AddIPToWhiteList(ctx, a.logger, ipData)
	if err != nil {
		message := helpers.StringBuild("AddIPToWhiteList IPData storage error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	message := helpers.StringBuild("IP added to whitelist(IP - ", ipData.IP, "/", strconv.Itoa(ipData.Mask), ")")
	a.logger.Info(message)

	return id, nil
}

func (a *App) RemoveIPInWhiteList(ctx context.Context, ipData storageData.StorageIPData) error {
	err := SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInWhiteList validate IPData error", err.Error())
		a.logger.Error(message)

		return err
	}
	err = a.storage.RemoveIPInWhiteList(ctx, a.logger, ipData)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInWhiteList app error(IP - ", ipData.IP, "),error: ", err.Error())
		a.logger.Error(message)

		return err
	}
	message := helpers.StringBuild("IP remove from whitelist(IP - ", ipData.IP, "/", strconv.Itoa(ipData.Mask), ")")
	a.logger.Info(message)

	return nil
}

func (a *App) IsIPInWhiteList(ctx context.Context, ipData storageData.StorageIPData) (bool, error) {
	err := SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("IsIPInWhiteList validate IPData error", err.Error())
		a.logger.Error(message)

		return false, err
	}
	ok, err := a.storage.IsIPInWhiteList(ctx, a.logger, ipData)

	return ok, err
}

func (a *App) GetAllIPInWhiteList(ctx context.Context) ([]storageData.StorageIPData, error) {
	whiteList, err := a.storage.GetAllIPInWhiteList(ctx, a.logger)

	return whiteList, err
}

func (a *App) isIPInWhiteListCheck(ctx context.Context, ip string) (bool, error) {
	canIP := net.ParseIP(ip)
	whiteList, err := a.storage.GetAllIPInWhiteList(ctx, a.logger)
	if err != nil {
		return false, err
	}
	for _, curWhiteIPData := range whiteList {
		curWhiteIPStr := curWhiteIPData.IP + "/" + strconv.Itoa(curWhiteIPData.Mask)
		_, subnet, err := net.ParseCIDR(curWhiteIPStr)
		if err != nil {
			return false, err
		}
		if subnet.Contains(canIP) {
			return true, nil
		}
	}

	return false, nil
}

// BLACKLIST

func (a *App) AddIPToBlackList(ctx context.Context, IPData storageData.StorageIPData) (int, error) {
	err := SimpleIPDataValidator(IPData, false)
	if err != nil {
		message := helpers.StringBuild("AddIPToBlackList validate IPData error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	ok, err := a.storage.IsIPInWhiteList(ctx, a.logger, IPData)
	if err != nil {
		message := helpers.StringBuild("AddIPToBlackList validate in whitelist IPData error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	if ok {
		return 0, ErrIPDataExistInWL
	}
	id, err := a.storage.AddIPToBlackList(ctx, a.logger, IPData)
	if err != nil {
		message := helpers.StringBuild("AddIPToBlackList IPData storage error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	message := helpers.StringBuild("IP added to blacklist(IP - ", IPData.IP, "/", strconv.Itoa(IPData.Mask), ")")
	a.logger.Info(message)

	return id, nil
}

func (a *App) RemoveIPInBlackList(ctx context.Context, ipData storageData.StorageIPData) error {
	err := SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInBlackList validate IPData error", err.Error())
		a.logger.Error(message)

		return err
	}
	err = a.storage.RemoveIPInBlackList(ctx, a.logger, ipData)
	if err != nil {
		errBaseText := "RemoveIPInBlackList app error(IP - "
		message := helpers.StringBuild(errBaseText, ipData.IP, "/", strconv.Itoa(ipData.Mask), "),error: ", err.Error())
		a.logger.Error(message)

		return err
	}
	message := helpers.StringBuild("IP remove from blacklist(IP - ", ipData.IP, ")")
	a.logger.Info(message)

	return nil
}

func (a *App) IsIPInBlackList(ctx context.Context, ipData storageData.StorageIPData) (bool, error) {
	err := SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("IsIPInBlackList validate IPData error", err.Error())
		a.logger.Error(message)

		return false, err
	}
	ok, err := a.storage.IsIPInBlackList(ctx, a.logger, ipData)

	return ok, err
}

func (a *App) GetAllIPInBlackList(ctx context.Context) ([]storageData.StorageIPData, error) {
	whiteList, err := a.storage.GetAllIPInBlackList(ctx, a.logger)

	return whiteList, err
}

func (a *App) isIPInBlackListCheck(ctx context.Context, ip string) (bool, error) {
	canIP := net.ParseIP(ip)
	blackList, err := a.storage.GetAllIPInBlackList(ctx, a.logger)
	if err != nil {
		return false, err
	}
	for _, curBlackIPData := range blackList {
		curBlackIPStr := curBlackIPData.IP + "/" + strconv.Itoa(curBlackIPData.Mask)
		_, subnet, err := net.ParseCIDR(curBlackIPStr)
		if err != nil {
			return false, err
		}
		if subnet.Contains(canIP) {
			return true, nil
		}
	}

	return false, nil
}
