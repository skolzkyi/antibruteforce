package app

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
	"go.uber.org/zap"
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
	AddIPToList(ctx context.Context, listname string, logger storageData.Logger, IPData storageData.StorageIPData) (int, error) //nolint: lll, nolintlint
	RemoveIPInList(ctx context.Context, listname string, logger storageData.Logger, IPData storageData.StorageIPData) error     //nolint: lll, nolintlint
	IsIPInList(ctx context.Context, listname string, logger storageData.Logger, IPData storageData.StorageIPData) (bool, error) //nolint: lll, nolintlint
	GetAllIPInList(ctx context.Context, listname string, logger storageData.Logger) ([]storageData.StorageIPData, error)
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
	ok, err := a.isIPInListCheck(ctx, "blacklist", req.IP)
	if err != nil {
		message := helpers.StringBuild("CheckInputRequest isIPInListCheck error: ", err.Error())
		a.logger.Error(message)

		return false, "", err
	}
	if ok {
		return false, "IP in blacklist", nil
	}
	ok, err = a.isIPInListCheck(ctx, "whitelist", req.IP)
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
		errBaseText := "CheckInputRequest IncrementAndGetBucketValue - IP error: "
		message := helpers.StringBuild(errBaseText, err.Error(), ", key: ", "ip_"+req.IP)
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

func (a *App) AddIPToList(ctx context.Context, listname string, ipData storageData.StorageIPData) (int, error) { //nolint: dupl, nolintlint
	err := SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("AddIPToList validate IPData error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	var otherlistname string
	switch listname {
	case storageData.WhiteListName:
		otherlistname = storageData.BlackListName
	case storageData.BlackListName:
		otherlistname = storageData.WhiteListName
	default:
		return 0, storageData.ErrErrorBadListType
	}

	ok, err := a.storage.IsIPInList(ctx, otherlistname, a.logger, ipData)
	if err != nil {
		message := helpers.StringBuild("AddIPToList validate in otherlist IPData error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	if ok {
		switch listname {
		case storageData.WhiteListName:
			return 0, ErrIPDataExistInBL
		case storageData.BlackListName:
			return 0, ErrIPDataExistInWL
		default:
			return 0, storageData.ErrErrorBadListType
		}
	}
	id, err := a.storage.AddIPToList(ctx, listname, a.logger, ipData)
	if err != nil {
		message := helpers.StringBuild("AddIPToList IPData storage error", err.Error())
		a.logger.Error(message)

		return 0, err
	}
	message := helpers.StringBuild("IP added to ", listname, "(IP - ", ipData.IP, "/", strconv.Itoa(ipData.Mask), ")")
	a.logger.Info(message)

	return id, nil
}

func (a *App) RemoveIPInList(ctx context.Context, listname string, ipData storageData.StorageIPData) error { //nolint: dupl, nolintlint
	err := checkListnName(listname)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInList checkListnName error", err.Error())
		a.logger.Error(message)
		return err
	}
	err = SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInList validate IPData error", err.Error())
		a.logger.Error(message)

		return err
	}
	err = a.storage.RemoveIPInList(ctx, listname, a.logger, ipData)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInList app error(IP - ", ipData.IP, "),error: ", err.Error())
		a.logger.Error(message)

		return err
	}
	message := helpers.StringBuild("IP remove from "+listname+"(IP - ", ipData.IP, "/", strconv.Itoa(ipData.Mask), ")")
	a.logger.Info(message)

	return nil
}

func (a *App) IsIPInList(ctx context.Context, listname string, ipData storageData.StorageIPData) (bool, error) { //nolint: dupl, nolintlint
	err := checkListnName(listname)
	if err != nil {
		message := helpers.StringBuild("IsIPInList  checkListnName error", err.Error())
		a.logger.Error(message)
		return false, err
	}
	err = SimpleIPDataValidator(ipData, false)
	if err != nil {
		message := helpers.StringBuild("IsIPInList validate IPData error", err.Error())
		a.logger.Error(message)
		return false, err
	}
	ok, err := a.storage.IsIPInList(ctx, listname, a.logger, ipData)

	return ok, err
}

func (a *App) GetAllIPInList(ctx context.Context, listname string) ([]storageData.StorageIPData, error) { //nolint: dupl, nolintlint
	err := checkListnName(listname)
	if err != nil {
		message := helpers.StringBuild("GetAllIPInList  checkListnName error", err.Error())
		a.logger.Error(message)
		return nil, err
	}
	list, err := a.storage.GetAllIPInList(ctx, listname, a.logger)

	return list, err
}

func (a *App) isIPInListCheck(ctx context.Context, listname string, ip string) (bool, error) { //nolint: dupl, nolintlint
	err := checkListnName(listname)
	if err != nil {
		message := helpers.StringBuild("isIPInListCheck checkListnName error", err.Error())
		a.logger.Error(message)
		return false, err
	}
	canIP := net.ParseIP(ip)
	list, err := a.storage.GetAllIPInList(ctx, listname, a.logger)
	if err != nil {
		return false, err
	}
	for _, curIPData := range list {
		curIPStr := curIPData.IP + "/" + strconv.Itoa(curIPData.Mask)
		_, subnet, err := net.ParseCIDR(curIPStr)
		if err != nil {
			return false, err
		}
		if subnet.Contains(canIP) {
			return true, nil
		}
	}

	return false, nil
}

func checkListnName(listname string) error {
	if listname != storageData.WhiteListName && listname != storageData.BlackListName {
		return storageData.ErrErrorBadListType
	}

	return nil
}
