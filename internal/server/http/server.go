package internalhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	storagedIP "github.com/skolzkyi/hwOTUS_YIA/hw12_13_14_15_calendar/internal/storage/storagedIP"
	"go.uber.org/zap"
)

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

type Application interface {
	InitStorage(ctx context.Context, config storage.Config) error
	CloseStorage(ctx context.Context) error
	AddIPToWhiteList(ctx context.Context, IP string) (int, error)
	RemoveIPInWhiteList(ctx context.Context, IP string) error
	IsIPInWhiteList(ctx context.Context, IP string) (bool, error)
	GetAllIPInWhiteList(ctx context.Context) ([]storagedIP.storagedIP, error)
	AddIPToBlackList(ctx context.Context, IP string) (int, error)
	RemoveIPInBlackList(ctx context.Context, IP string) error
	IsIPInBlackList(ctx context.Context, IP string) (bool, error)
	GetAllIPInBlackList(ctx context.Context) ([]storagedIP.storagedIP, error)
}

func NewServer(logger Logger, app Application, config Config) *Server {
	server := Server{}
	server.logg = logger
	server.app = app
	server.Config = config
	server.serv = &http.Server{
		Addr:    config.GetServerURL(),
		Handler: server.routes(),
		ReadHeaderTimeout: 2 * time.Second,
	}

	return &server
}

func (s *Server) Start(ctx context.Context) error {
	s.logg.Info("calendar is running...")
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
	s.logg.Info("server graceful shutdown")
	return err
}
