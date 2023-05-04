//go:build !integration
// +build !integration

package app

import (
	"context"
	"errors"
	"testing"
	"time"

	logger "github.com/skolzkyi/antibruteforce/internal/logger"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
	storageSQLMock "github.com/skolzkyi/antibruteforce/internal/storage/storageSQLMock"
	"github.com/stretchr/testify/require"
)

type ConfigTest struct{}

func (config *ConfigTest) Init(_ string) error {
	return nil
}

func (config *ConfigTest) GetServerURL() string {
	return "127.0.0.1:4000"
}

func (config *ConfigTest) GetAddress() string {
	return "127.0.0.1"
}

func (config *ConfigTest) GetPort() string {
	return "4000"
}

func (config *ConfigTest) GetServerShutdownTimeout() time.Duration {
	return 5 * time.Second
}

func (config *ConfigTest) GetDBName() string {
	return "OTUSAntibruteforce"
}

func (config *ConfigTest) GetDBUser() string {
	return "imapp"
}

func (config *ConfigTest) GetDBPassword() string {
	return "LightInDark"
}

func (config *ConfigTest) GetDBConnMaxLifetime() time.Duration {
	return 5 * time.Second
}

func (config *ConfigTest) GetDBMaxOpenConns() int {
	return 20
}

func (config *ConfigTest) GetDBMaxIdleConns() int {
	return 20
}

func (config *ConfigTest) GetDBTimeOut() time.Duration {
	return 5 * time.Second
}

func (config *ConfigTest) GetDBAddress() string {
	return "127.0.0.1"
}

func (config *ConfigTest) GetDBPort() string {
	return "3306"
}

func (config *ConfigTest) GetResisAddress() string {
	return "127.0.0.1"
}

func (config *ConfigTest) GetRedisPort() string {
	return "6379"
}

func (config *ConfigTest) GetLimitFactorLogin() int{
	return 10
}

func (config *ConfigTest) GetLimitFactorPassword() int {
	return 100
}

func (config *ConfigTest) GetLimitFactorIP() int{
	return 1000
}

func (config *ConfigTest) GetLimitTimeCheck() time.Duration {
	return 60 * time.Second
}

func initAppWithMocks(t *testing.T) *App {
	t.Helper()
	logger, _ := logger.New("debug")
	config:=ConfigTest{}
	var storage Storage
	storage = storageSQLMock.New()
	ctxStor, cancelStore := context.WithTimeout(context.Background(), config.GetDBTimeOut())
	err := storage.Init(ctxStor, logger, &config)
	if err != nil {
		cancelStore()
	}
	require.NoError(t, err)
	require.NoError(t, err)
	antibruteforce := New(logger, storage)
	return antibruteforce
}

func TestSimpleRequestValidator(t *testing.T) {
	t.Run("PositiveRequestValidator", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("user0","root","192.168.0.12") 
		require.NoError(t, err)
	})
	t.Run("NegativeErrVoidLogin", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("","root","192.168.0.12") 
		require.Truef(t, errors.Is(err, ErrVoidLogin), "actual error %q", err)
	})
	t.Run("NegativeErrVoidPassword", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("user0","","192.168.0.12") 
		require.Truef(t, errors.Is(err, ErrVoidPassword), "actual error %q", err)
	})
	t.Run("NegativeErrVoidIP", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("user0","root","") 
		require.Truef(t, errors.Is(err, ErrVoidIP), "actual error %q", err)
	})
}


func TestSimpleIPDataValidator(t *testing.T) {
	t.Run("PositiveIPDataValidator", func(t *testing.T) {
		t.Parallel()
		testData:=storageData.StorageIPData{
			IP:   "192.168.0.0",
			Mask: 25,
		}
		err := SimpleIPDataValidator(testData,false) 
		require.NoError(t, err)
	})
	t.Run("PositiveIPDataValidatorALL", func(t *testing.T) {
		t.Parallel()
		testData:=storageData.StorageIPData{
			IP:   "ALL",
			Mask: 0,
		}
		err := SimpleIPDataValidator(testData,true) 
		require.NoError(t, err)
	})
	t.Run("NegativeErrVoidIP", func(t *testing.T) {
		t.Parallel()
		testData:=storageData.StorageIPData{
			IP:   "",
			Mask: 25,
		}
		err := SimpleIPDataValidator(testData,false) 
		require.Truef(t, errors.Is(err, ErrVoidIP), "actual error %q", err)
	})
	t.Run("NegativeErrVoidMask", func(t *testing.T) {
		t.Parallel()
		testData:=storageData.StorageIPData{
			IP:   "192.168.0.0",
			Mask: 0,
		}
		err := SimpleIPDataValidator(testData,false) 
		require.Truef(t, errors.Is(err, ErrVoidIP), "actual error %q", err)
	})
}

func TestAppPositiveAddIPToWhiteListAndIsIPInWhiteList(t *testing.T) {
	app := initAppWithMocks(t)
	config:=ConfigTest{}
	err:=app.InitStorage(context.Background(),&config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newData:=storageData.StorageIPData{
		IP:   "192.168.0.0",
		Mask: 25,
	}
	_, err = app.AddIPToWhiteList(context.Background(), newData)
	require.NoError(t, err) 
	ok, err := app.IsIPInWhiteList(context.Background(), newData)
	require.NoError(t, err) 
	require.Truef(t, ok == true, "IP not in whitelist", ok)
}

func TestAppPositiveRemoveIPInWhiteListAndIsIPInWhiteList(t *testing.T) {
	app := initAppWithMocks(t)
	config:=ConfigTest{}
	err:=app.InitStorage(context.Background(),&config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newData:=storageData.StorageIPData{
		IP:   "192.168.0.0",
		Mask: 25,
	}
	_, err = app.AddIPToWhiteList(context.Background(), newData)
	require.NoError(t, err) 
	ok, err := app.IsIPInWhiteList(context.Background(), newData)
	require.NoError(t, err) 
	require.Truef(t, ok == true, "IP not in whitelist", ok)
	err = app.RemoveIPInWhiteList(context.Background(), newData)
	require.NoError(t, err)
	ok, err = app.IsIPInWhiteList(context.Background(), newData)
	require.NoError(t, err) 
	require.Truef(t, ok == false, "IP in whitelist after removing", ok)
}

func TestAppPositiveGetAllIPInWhiteList(t *testing.T) {
	app := initAppWithMocks(t)
	config:=ConfigTest{}
	err:=app.InitStorage(context.Background(),&config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newDataSl:=make([]storageData.StorageIPData,2)
	newDataSl[0]=storageData.StorageIPData{
		ID:   0,
		IP:   "192.168.0.0",
		Mask: 25,
	}
	newDataSl[1]=storageData.StorageIPData{
		ID:   1,
		IP:   "10.0.0.0",
		Mask: 8,
	}
    for _,curData:= range newDataSl {
		_, err = app.AddIPToWhiteList(context.Background(), curData)
		require.NoError(t, err) 
	}
	
	controlDataSl,err:=app.GetAllIPInWhiteList(context.Background())
	require.NoError(t, err) 
	require.Equal(t,newDataSl,controlDataSl)
}
