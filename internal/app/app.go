package app

import (
	"context"
	"strconv"
	//"time"
	"go.uber.org/zap"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
    storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)



type App struct {
	logger 		  Logger
	storage       Storage
	//bucketStorage BStorage
}

type Logger interface {
	Info(msg string)
	Warning(msg string)
	Error(msg string)
	Fatal(msg string)
	GetZapLogger() *zap.SugaredLogger
}
/*
type storageIPData struct {
	IP                    string
	ID                    int
}
*/
type Storage interface {
	Init(ctx context.Context, logger storageData.Logger, config storageData.Config) error
	Close(ctx context.Context, logger storageData.Logger) error
	AddIPToWhiteList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData)(int, error) 
	RemoveIPInWhiteList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) error 
	IsIPInWhiteList(ctx context.Context,logger storageData.Logger, IPData storageData.StorageIPData) (bool,error)
	GetAllIPInWhiteList(ctx context.Context,logger storageData.Logger) ([]storageData.StorageIPData,error)
	AddIPToBlackList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData)(int, error) 
	RemoveIPInBlackList(ctx context.Context, logger storageData.Logger, IPData storageData.StorageIPData) error 
	IsIPInBlackList(ctx context.Context,logger storageData.Logger, IPData storageData.StorageIPData) (bool,error) 
	GetAllIPInBlackList(ctx context.Context,logger storageData.Logger) ([]storageData.StorageIPData,error)
}

type BStorage interface {
	Init(ctx context.Context, logger storageData.Logger, config storageData.Config) error
	Close(ctx context.Context, logger storageData.Logger) error
	IncrementBucketValue(ctx context.Context, logger storageData.Logger, key string) error
	GetBucketValue(ctx context.Context, logger storageData.Logger, key string) (int,error)
	FlushBStorage(ctx context.Context, logger storageData.Logger) error
}

func New(logger Logger, storage Storage) *App {
	app := App{
		logger:        logger,
		storage:       storage,
	//	bucketStorage: bStorage,
	}
	return &app
}

func (a *App) InitStorage(ctx context.Context, config storageData.Config) error {
	return a.storage.Init(ctx, a.logger, config)
}

func (a *App) CloseStorage(ctx context.Context) error {
	return a.storage.Close(ctx, a.logger)
}

// WHITELIST

func (a *App) AddIPToWhiteList(ctx context.Context, IPData storageData.StorageIPData) (int, error) {
	err := SimpleIPDataValidator(IPData,false)
	if err != nil {
		message := helpers.StringBuild("AddIPToWhiteList validate IPData error", err.Error())
		a.logger.Error(message)
		return 0,err
	}
	id, err := a.storage.AddIPToWhiteList(ctx, a.logger, IPData)
	if err != nil {
		message := helpers.StringBuild("AddIPToWhiteList IPData storage error", err.Error())
		a.logger.Error(message)
		return 0,err
	}
	message := helpers.StringBuild("IP added to whitelist(IP - ", IPData.IP,"/",strconv.Itoa(IPData.Mask), ")")
	a.logger.Info(message)
	return id, nil
}

func (a *App) RemoveIPInWhiteList(ctx context.Context, IPData storageData.StorageIPData) error {
	err := SimpleIPDataValidator(IPData,false)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInWhiteList validate IPData error", err.Error())
		a.logger.Error(message)
		return err
	}
	err = a.storage.RemoveIPInWhiteList(ctx, a.logger, IPData)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInWhiteList app error(IP - ", IPData.IP, "),error: ", err.Error())
		a.logger.Error(message)
		return err
	}
	message := helpers.StringBuild("IP remove from whitelist(IP - ", IPData.IP,"/",strconv.Itoa(IPData.Mask), ")")
	a.logger.Info(message)

	return nil
}

func (a *App) IsIPInWhiteList(ctx context.Context, IPData storageData.StorageIPData) (bool, error) {
	err := SimpleIPDataValidator(IPData,false)
	if err != nil {
		message := helpers.StringBuild("IsIPInWhiteList validate IPData error", err.Error())
		a.logger.Error(message)
		return false,err
	}
	ok, err := a.storage.IsIPInWhiteList(ctx, a.logger, IPData)
	return ok, err
}

func (a *App) GetAllIPInWhiteList(ctx context.Context) ([]storageData.StorageIPData, error) {
	whiteList, err := a.storage.GetAllIPInWhiteList(ctx, a.logger)
	return whiteList, err
}

// BLACKLIST 

func (a *App) AddIPToBlackList(ctx context.Context, IPData storageData.StorageIPData) (int, error) {
	err := SimpleIPDataValidator(IPData,false)
	if err != nil {
		message := helpers.StringBuild("AddIPToBlackList validate IPData error", err.Error())
		a.logger.Error(message)
		return 0,err
	}
	id, err := a.storage.AddIPToBlackList(ctx, a.logger, IPData)
	if err != nil {
		message := helpers.StringBuild("AddIPToBlackList IPData storage error", err.Error())
		a.logger.Error(message)
		return 0,err
	}
	message := helpers.StringBuild("IP added to blacklist(IP - ", IPData.IP,"/",strconv.Itoa(IPData.Mask), ")")
	a.logger.Info(message)
	return id, nil
}

func (a *App) RemoveIPInBlackList(ctx context.Context, IPData storageData.StorageIPData) error {
	err := SimpleIPDataValidator(IPData,false)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInBlackList validate IPData error", err.Error())
		a.logger.Error(message)
		return err
	}
	err = a.storage.RemoveIPInBlackList(ctx, a.logger, IPData)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInBlackList app error(IP - ", IPData.IP,"/",strconv.Itoa(IPData.Mask),"),error: ", err.Error())
		a.logger.Error(message)
		return err
	}
	message := helpers.StringBuild("IP remove from blacklist(IP - ", IPData.IP, ")")
	a.logger.Info(message)

	return nil
}

func (a *App) IsIPInBlackList(ctx context.Context, IPData storageData.StorageIPData) (bool, error) {
	err := SimpleIPDataValidator(IPData,false)
	if err != nil {
		message := helpers.StringBuild("IsIPInBlackList validate IPData error", err.Error())
		a.logger.Error(message)
		return false, err
	}
	ok, err := a.storage.IsIPInBlackList(ctx, a.logger, IPData)
	return ok, err
}

func (a *App) GetAllIPInBlackList(ctx context.Context) ([]storageData.StorageIPData, error) {
	whiteList, err := a.storage.GetAllIPInBlackList(ctx, a.logger)
	return whiteList, err
}
