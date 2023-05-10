//go:build !integration
// +build !integration

package app

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	logger "github.com/skolzkyi/antibruteforce/internal/logger"
	RedisStorage "github.com/skolzkyi/antibruteforce/internal/storage/redis"
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

func (config *ConfigTest) GetRedisAddress() string {
	return "127.0.0.1"
}

func (config *ConfigTest) GetRedisPort() string {
	return "6379"
}

func (config *ConfigTest) GetLimitFactorLogin() int {
	return 10
}

func (config *ConfigTest) GetLimitFactorPassword() int {
	return 100
}

func (config *ConfigTest) GetLimitFactorIP() int {
	return 12
}

func (config *ConfigTest) GetLimitTimeCheck() time.Duration {
	return 60 * time.Second
}

func initAppWithMocks(t *testing.T) *App {
	t.Helper()
	logger, _ := logger.New("debug")
	config := ConfigTest{}
	var storage Storage
	storage = storageSQLMock.New()
	ctxStor, _ := context.WithTimeout(context.Background(), config.GetDBTimeOut())
	err := storage.Init(ctxStor, logger, &config)
	require.NoError(t, err)
	redis := RedisStorage.New()
	err = redis.InitAsMock(ctxStor, logger)
	require.NoError(t, err)
	antibruteforce := New(logger, storage, redis, &config)
	return antibruteforce
}

func TestSimpleRequestValidator(t *testing.T) {
	t.Run("PositiveRequestValidator", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("user0", "root", "192.168.0.12")
		require.NoError(t, err)
	})
	t.Run("NegativeErrVoidLogin", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("", "root", "192.168.0.12")
		require.Truef(t, errors.Is(err, ErrVoidLogin), "actual error %q", err)
	})
	t.Run("NegativeErrVoidPassword", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("user0", "", "192.168.0.12")
		require.Truef(t, errors.Is(err, ErrVoidPassword), "actual error %q", err)
	})
	t.Run("NegativeErrVoidIP", func(t *testing.T) {
		t.Parallel()
		_, err := SimpleRequestValidator("user0", "root", "")
		require.Truef(t, errors.Is(err, ErrBadIP), "actual error %q", err)
	})
}

func TestSimpleIPDataValidator(t *testing.T) {
	t.Run("PositiveIPDataValidator", func(t *testing.T) {
		t.Parallel()
		testData := storageData.StorageIPData{
			IP:   "192.168.0.0",
			Mask: 25,
		}
		err := SimpleIPDataValidator(testData, false)
		require.NoError(t, err)
	})
	t.Run("PositiveIPDataValidatorALL", func(t *testing.T) {
		t.Parallel()
		testData := storageData.StorageIPData{
			IP:   "ALL",
			Mask: 0,
		}
		err := SimpleIPDataValidator(testData, true)
		require.NoError(t, err)
	})
	t.Run("NegativeErrVoidIP", func(t *testing.T) {
		t.Parallel()
		testData := storageData.StorageIPData{
			IP:   "",
			Mask: 25,
		}
		err := SimpleIPDataValidator(testData, false)
		require.Truef(t, errors.Is(err, ErrBadIP), "actual error %q", err)
	})
	t.Run("NegativeErrVoidMask", func(t *testing.T) {
		t.Parallel()
		testData := storageData.StorageIPData{
			IP:   "192.168.0.0",
			Mask: 0,
		}
		err := SimpleIPDataValidator(testData, false)
		require.Truef(t, errors.Is(err, ErrVoidMask), "actual error %q", err)
	})
}

func TestAppNegativeAddIPCrossAdding(t *testing.T) {
	app := initAppWithMocks(t)
	config := ConfigTest{}
	err := app.InitStorage(context.Background(), &config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newData := storageData.StorageIPData{
		IP:   "192.168.0.0",
		Mask: 25,
	}
	_, err = app.AddIPToWhiteList(context.Background(), newData)
	require.NoError(t, err)
	ok, err := app.IsIPInWhiteList(context.Background(), newData)
	require.NoError(t, err)
	require.Truef(t, ok == true, "IP not in whitelist", ok)
	_, err = app.AddIPToBlackList(context.Background(), newData)
	require.Truef(t, errors.Is(err, ErrIPDataExistInWL), "actual error %q", err)
	newData = storageData.StorageIPData{
		IP:   "10.0.0.0",
		Mask: 8,
	}
	_, err = app.AddIPToBlackList(context.Background(), newData)
	require.NoError(t, err)
	ok, err = app.IsIPInBlackList(context.Background(), newData)
	require.NoError(t, err)
	require.Truef(t, ok == true, "IP not in blacklist", ok)
	_, err = app.AddIPToWhiteList(context.Background(), newData)
	require.Truef(t, errors.Is(err, ErrIPDataExistInBL), "actual error %q", err)
}

// WHITELIST

func TestAppPositiveAddIPToWhiteListAndIsIPInWhiteList(t *testing.T) {
	app := initAppWithMocks(t)
	config := ConfigTest{}
	err := app.InitStorage(context.Background(), &config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newData := storageData.StorageIPData{
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
	config := ConfigTest{}
	err := app.InitStorage(context.Background(), &config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newData := storageData.StorageIPData{
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
	config := ConfigTest{}
	err := app.InitStorage(context.Background(), &config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newDataSl := make([]storageData.StorageIPData, 2)
	newDataSl[0] = storageData.StorageIPData{
		ID:   0,
		IP:   "192.168.0.0",
		Mask: 25,
	}
	newDataSl[1] = storageData.StorageIPData{
		ID:   1,
		IP:   "10.0.0.0",
		Mask: 8,
	}
	for _, curData := range newDataSl {
		_, err = app.AddIPToWhiteList(context.Background(), curData)
		require.NoError(t, err)
	}

	controlDataSl, err := app.GetAllIPInWhiteList(context.Background())
	require.NoError(t, err)
	require.Equal(t, newDataSl, controlDataSl)
}

// BLACKLIST

func TestAppPositiveAddIPToBlackListAndIsIPInBlackList(t *testing.T) {
	app := initAppWithMocks(t)
	config := ConfigTest{}
	err := app.InitStorage(context.Background(), &config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newData := storageData.StorageIPData{
		IP:   "192.168.0.0",
		Mask: 25,
	}
	_, err = app.AddIPToBlackList(context.Background(), newData)
	require.NoError(t, err)
	ok, err := app.IsIPInBlackList(context.Background(), newData)
	require.NoError(t, err)
	require.Truef(t, ok == true, "IP not in blacklist", ok)
}

func TestAppPositiveRemoveIPInBlackListAndIsIPInBlackList(t *testing.T) {
	app := initAppWithMocks(t)
	config := ConfigTest{}
	err := app.InitStorage(context.Background(), &config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newData := storageData.StorageIPData{
		IP:   "192.168.0.0",
		Mask: 25,
	}
	_, err = app.AddIPToBlackList(context.Background(), newData)
	require.NoError(t, err)
	ok, err := app.IsIPInBlackList(context.Background(), newData)
	require.NoError(t, err)
	require.Truef(t, ok == true, "IP not in whitelist", ok)
	err = app.RemoveIPInBlackList(context.Background(), newData)
	require.NoError(t, err)
	ok, err = app.IsIPInBlackList(context.Background(), newData)
	require.NoError(t, err)
	require.Truef(t, ok == false, "IP in whitelist after removing", ok)
}

func TestAppPositiveGetAllIPInBlackList(t *testing.T) {
	app := initAppWithMocks(t)
	config := ConfigTest{}
	err := app.InitStorage(context.Background(), &config)
	require.NoError(t, err)
	defer app.CloseStorage(context.Background())
	newDataSl := make([]storageData.StorageIPData, 2)
	newDataSl[0] = storageData.StorageIPData{
		ID:   0,
		IP:   "192.168.0.0",
		Mask: 25,
	}
	newDataSl[1] = storageData.StorageIPData{
		ID:   1,
		IP:   "10.0.0.0",
		Mask: 8,
	}
	for _, curData := range newDataSl {
		_, err = app.AddIPToBlackList(context.Background(), curData)
		require.NoError(t, err)
	}

	controlDataSl, err := app.GetAllIPInBlackList(context.Background())
	require.NoError(t, err)
	require.Equal(t, newDataSl, controlDataSl)
}

// REQUEST AUTH

func TestRequestAuth(t *testing.T) {
	t.Run("PositiveRequestAuth", func(t *testing.T) {
		t.Parallel()
		app := initAppWithMocks(t)
		req := storageData.RequestAuth{
			Login:    "user0",
			Password: "CharlyDonTSerf",
			IP:       "192.168.16.56",
		}
		ok, message, err := app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "clear check", message)
	})

	t.Run("PositiveRequestAuthInWhiteList", func(t *testing.T) {
		t.Parallel()
		app := initAppWithMocks(t)
		req := storageData.RequestAuth{
			Login:    "user0",
			Password: "CharlyDonTSerf",
			IP:       "192.168.16.56",
		}
		newData := storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 24,
		}
		_, err := app.AddIPToWhiteList(context.Background(), newData)
		require.NoError(t, err)
		ok, message, err := app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "IP in whitelist", message)
	})
	t.Run("PositiveRequestAuthInBlackList", func(t *testing.T) {
		t.Parallel()
		app := initAppWithMocks(t)
		req := storageData.RequestAuth{
			Login:    "user0",
			Password: "CharlyDonTSerf",
			IP:       "192.168.16.56",
		}
		newData := storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 24,
		}
		_, err := app.AddIPToBlackList(context.Background(), newData)
		require.NoError(t, err)
		ok, message, err := app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Equal(t, "IP in blacklist", message)
	})
	t.Run("PositiveRequestAuthRateLimitByTag", func(t *testing.T) {
		t.Parallel()
		app := initAppWithMocks(t)
		req := storageData.RequestAuth{
			Login:    "user0",
			Password: "CharlyDonTSerf",
			IP:       "192.168.16.56",
		}
		for i := 0; i < 10; i++ {
			ok, message, err := app.CheckInputRequest(context.Background(), req)
			require.NoError(t, err)
			require.Equal(t, true, ok)
			require.Equal(t, "clear check", message)
		}
		ok, message, err := app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Equal(t, "rate limit by login", message)
	})
	t.Run("PositiveRequestAuthRateLimitByTagAndClearLoginBucket", func(t *testing.T) {
		t.Parallel()
		app := initAppWithMocks(t)
		req := storageData.RequestAuth{
			Login:    "user0",
			Password: "CharlyDonTSerf",
			IP:       "192.168.16.56",
		}
		for i := 0; i < 10; i++ {
			ok, message, err := app.CheckInputRequest(context.Background(), req)
			require.NoError(t, err)
			require.Equal(t, true, ok)
			require.Equal(t, "clear check", message)
		}
		ok, message, err := app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Equal(t, "rate limit by login", message)
		err = app.ClearBucketByLogin(context.Background(), "user0")
		require.NoError(t, err)
		ok, message, err = app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "clear check", message)
	})
	t.Run("PositiveRequestAuthRateLimitByTagAndClearLoginBucket", func(t *testing.T) {
		t.Parallel()
		app := initAppWithMocks(t)
		req := storageData.RequestAuth{
			Login:    "user0",
			Password: "CharlyDonTSerf",
			IP:       "192.168.16.56",
		}
		for i := 0; i < 12; i++ {
			req := storageData.RequestAuth{
				Login:    strconv.Itoa(i),
				Password: "CharlyDonTSerf",
				IP:       "192.168.16.56",
			}
			ok, message, err := app.CheckInputRequest(context.Background(), req)
			require.NoError(t, err)
			require.Equal(t, true, ok)
			require.Equal(t, "clear check", message)
		}
		ok, message, err := app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Equal(t, "rate limit by IP", message)
		err = app.ClearBucketByIP(context.Background(), "192.168.16.56")
		require.NoError(t, err)
		ok, message, err = app.CheckInputRequest(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "clear check", message)
	})
}
