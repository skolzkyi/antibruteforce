package app

import (
	"context"
	"strconv"
	"time"
	"go.uber.org/zap"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storagedIP "github.com/skolzkyi/antibruteforce/internal/storage/storagedIP"
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

type Storage interface {
	Init(ctx context.Context, logger storage.Logger, config storage.Config) error
	Close(ctx context.Context, logger storage.Logger) error
	AddIPToWhiteList(ctx context.Context, logger storage.Logger, IP string)(int, error) 
	RemoveIPInWhiteList(ctx context.Context, logger storage.Logger, IP string) error 
	IsIPInWhiteList(ctx context.Context,logger storage.Logger, IP string) (bool,error)
	GetAllIPInWhiteList(ctx context.Context,logger storage.Logger) ([]storagedIP.storagedIP,error)
	AddIPToBlackList(ctx context.Context, logger storage.Logger, IP string)(int, error) 
	RemoveIPInBlackList(ctx context.Context, logger storage.Logger, IP string) error 
	IsIPInBlackList(ctx context.Context,logger storage.Logger, IP string) (bool,error) 
	GetAllIPInBlackList(ctx context.Context,logger storage.Logger) ([]storagedIP.storagedIP,error)
}

type BStorage interface {
	Init(ctx context.Context, logger storage.Logger, config storage.Config) error
	Close(ctx context.Context, logger storage.Logger) error
	IncrementBucketValue(ctx context.Context, logger storage.Logger, key string) error
	GetBucketValue(ctx context.Context, logger storage.Logger, key string) (int,error)
	FlushBStorage(ctx context.Context, logger storage.Logger) error
}

func New(logger Logger, storage Storage) *App {
	app := App{
		logger:        logger,
		storage:       storage,
	//	bucketStorage: bStorage,
	}
	return &app
}

func (a *App) InitStorage(ctx context.Context, config storage.Config) error {
	return a.storage.Init(ctx, a.logger, config)
}

func (a *App) CloseStorage(ctx context.Context) error {
	return a.storage.Close(ctx, a.logger)
}

func (a *App) AddIPToWhiteList(ctx context.Context, IP string) (int, error) {
	id, err := a.storage.AddIPToWhiteList(ctx, a.logger, IP)
	return id, err
}

func (a *App) RemoveIPInWhiteList(ctx context.Context, IP string) error {
	err := a.storage.RemoveIPInWhiteList(ctx, a.logger, IP)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInWhiteList app error(IP - ", IP, "),error: ", err.Error())
		a.logger.Error(message)
		return err
	}
	message := helpers.StringBuild("IP remove from whitelist(IP - ", IP, ")")
	a.logger.Info(message)

	return nil
}

func (a *App) IsIPInWhiteList(ctx context.Context, IP string) (bool, error) {
	ok, err := a.storage.IsIPInWhiteList(ctx, a.logger, IP)
	return ok, err
}

func (a *App) GetAllIPInWhiteList(ctx context.Context) ([]storagedIP.storagedIP, error) {
	whiteList, err := a.storage.GetAllIPInWhiteList(ctx, a.logger)
	return whiteList, err
}

func (a *App) AddIPToBlackList(ctx context.Context, IP string) (int, error) {
	id, err := a.storage.AddIPToBlackList(ctx, a.logger, IP)
	return id, err
}

func (a *App) RemoveIPInBlackList(ctx context.Context, IP string) error {
	err := a.storage.RemoveIPInBlackList(ctx, a.logger, IP)
	if err != nil {
		message := helpers.StringBuild("RemoveIPInBlackList app error(IP - ", IP, "),error: ", err.Error())
		a.logger.Error(message)
		return err
	}
	message := helpers.StringBuild("IP remove from blacklist(IP - ", IP, ")")
	a.logger.Info(message)

	return nil
}

func (a *App) IsIPInBlackList(ctx context.Context, IP string) (bool, error) {
	ok, err := a.storage.IsIPInBlackList(ctx, a.logger, IP)
	return ok, err
}

func (a *App) GetAllIPInBlackList(ctx context.Context) ([]storagedIP.storagedIP, error) {
	whiteList, err := a.storage.GetAllIPInBlackList(ctx, a.logger)
	return whiteList, err
}
