package internalhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
	"go.uber.org/zap"
)

var ErrBadBucketTypeTag = errors.New("bad bucket type tag")

type Server struct {
	serv   *http.Server
	logg   Logger
	app    Application
	Config Config
}

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

type Application interface {
	InitBStorageAndLimits(ctx context.Context, config storageData.Config) error
	CloseBStorage(ctx context.Context) error
	CheckInputRequest(ctx context.Context, req storageData.RequestAuth) (bool, string, error)
	RLTicker(ctx context.Context)
	ClearBucketByLogin(ctx context.Context, login string) error
	ClearBucketByIP(ctx context.Context, IP string) error
	InitStorage(ctx context.Context, config storageData.Config) error
	CloseStorage(ctx context.Context) error
	AddIPToList(ctx context.Context, listname string, IPData storageData.StorageIPData) (int, error)
	RemoveIPInList(ctx context.Context, listname string, IPData storageData.StorageIPData) error
	IsIPInList(ctx context.Context, listname string, IPData storageData.StorageIPData) (bool, error)
	GetAllIPInList(ctx context.Context, listname string) ([]storageData.StorageIPData, error)
}

func NewServer(logger Logger, app Application, config Config) *Server {
	server := Server{}
	server.logg = logger
	server.app = app
	server.Config = config
	server.serv = &http.Server{
		Addr:              config.GetServerURL(),
		Handler:           server.routes(),
		ReadHeaderTimeout: 2 * time.Second,
	}

	return &server
}

func (s *Server) Start(ctx context.Context) error {
	s.logg.Info("antibruteforce is running...")
	s.app.RLTicker(ctx)
	err := s.serv.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			s.logg.Error("server start error: " + err.Error())

			return err
		}
	}
	<-ctx.Done()

	return err
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.serv.Shutdown(ctx)
	if err != nil {
		s.logg.Error("server shutdown error: " + err.Error())

		return err
	}
	err = s.app.CloseStorage(ctx)
	if err != nil {
		s.logg.Error("server closeStorage error: " + err.Error())

		return err
	}
	err = s.app.CloseBStorage(ctx)
	if err != nil {
		s.logg.Error("server CloseBStorage error: " + err.Error())

		return err
	}
	s.logg.Info("server graceful shutdown")

	return err
}
