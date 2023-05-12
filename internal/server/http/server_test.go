//go:build !integration
// +build !integration

package internalhttp

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	app "github.com/skolzkyi/antibruteforce/internal/app"
	logger "github.com/skolzkyi/antibruteforce/internal/logger"
	RedisStorage "github.com/skolzkyi/antibruteforce/internal/storage/redis"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
	storageSQLMock "github.com/skolzkyi/antibruteforce/internal/storage/storageSQLMock"
	"github.com/stretchr/testify/require"
)

const correctOutputJSONAnswer string = `{"Text":"OK!","Code":0}`

const localhost string = "127.0.0.1"

type ConfigTest struct{}

func (config *ConfigTest) Init(_ string) error {
	return nil
}

func (config *ConfigTest) GetServerURL() string {
	return "127.0.0.1:4000"
}

func (config *ConfigTest) GetAddress() string {
	return localhost
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
	return localhost
}

func (config *ConfigTest) GetDBPort() string {
	return "3306"
}

func (config *ConfigTest) GetRedisAddress() string {
	return localhost
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

func TestWhiteListREST(t *testing.T) {
	t.Parallel()
	t.Run("AddIPToWhiteList", func(t *testing.T) {//nolint: dupl, nolintlint
		t.Parallel()
		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		r := httptest.NewRequest("POST", "/whitelist/", data)
		w := httptest.NewRecorder()
		server.WhiteList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := correctOutputJSONAnswer
		require.Equal(t, respExp, string(respBody))
	})
	t.Run("IsIPInWhiteList", func(t *testing.T) {//nolint: dupl, nolintlint
		t.Parallel()
		newData := storageData.StorageIPData{
			IP:   "192.168.16.1",
			Mask: 8,
		}

		data := bytes.NewBufferString(`{
			"IP":"192.168.16.1",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToList(context.Background(), "whitelist", newData)
		require.NoError(t, err)

		r := httptest.NewRequest("GET", "/whitelist/", data)
		w := httptest.NewRecorder()
		server.WhiteList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := `{"IPList":[],"Message":{"Text":"YES","Code":0}}`
		require.Equal(t, respExp, string(respBody))
	})
	t.Run("RemoveIPInWhiteList", func(t *testing.T) {//nolint: dupl, nolintlint
		t.Parallel()
		controldataTestIP := "192.168.16.0"
		newData := storageData.StorageIPData{
			IP:   controldataTestIP,
			Mask: 8,
		}

		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToList(context.Background(), "whitelist", newData)
		require.NoError(t, err)
		controlDataSl, err := server.app.GetAllIPInList(context.Background(), "whitelist")
		require.NoError(t, err)
		flag := false
		for _, curControlData := range controlDataSl {
			if curControlData.IP == controldataTestIP && curControlData.Mask == 8 {
				flag = true

				break
			}
		}
		require.Equal(t, flag, true)

		r := httptest.NewRequest("DELETE", "/whitelist/", data)
		w := httptest.NewRecorder()
		server.WhiteList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := correctOutputJSONAnswer
		require.Equal(t, respExp, string(respBody))

		controlDataSl, err = server.app.GetAllIPInList(context.Background(), "whitelist")
		require.NoError(t, err)
		flag = false
		for _, curControlData := range controlDataSl {
			if curControlData.IP == controldataTestIP && curControlData.Mask == 8 {
				flag = true

				break
			}
		}
		require.Equal(t, flag, false)
	})
	t.Run("GetAllIPInWhiteList", func(t *testing.T) {//nolint: dupl, nolintlint
		t.Parallel()
		data := bytes.NewBufferString(`{
			"IP":"ALL",
			"Mask":0
		
		}`)
		newData := storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		server := createServer(t)

		_, err := server.app.AddIPToList(context.Background(), "whitelist", newData)
		require.NoError(t, err)
		newData = storageData.StorageIPData{
			IP:   "172.92.24.0",
			Mask: 24,
		}
		_, err = server.app.AddIPToList(context.Background(), "whitelist", newData)
		require.NoError(t, err)

		r := httptest.NewRequest("GET", "/whitelist/", data)
		w := httptest.NewRecorder()
		server.WhiteList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := `{"IPList":[{"IP":"192.168.16.0","Mask":8,"ID":0},` +
			`{"IP":"172.92.24.0","Mask":24,"ID":1}],"Message":{"Text":"OK!","Code":0}}`
		require.Equal(t, respExp, string(respBody))
	})
}

func TestBlackListREST(t *testing.T) {
	t.Parallel()
	t.Run("AddIPToBlackList", func(t *testing.T) {//nolint: dupl, nolintlint
		t.Parallel()
		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		r := httptest.NewRequest("POST", "/blacklist/", data)
		w := httptest.NewRecorder()
		server.BlackList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := correctOutputJSONAnswer
		require.Equal(t, respExp, string(respBody))
	})
	t.Run("IsIPInBlackList", func(t *testing.T) {
		t.Parallel()
		newData := storageData.StorageIPData{//nolint: dupl, nolintlint
			IP:   "192.168.16.0",
			Mask: 8,
		}

		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToList(context.Background(), "blacklist", newData)
		require.NoError(t, err)

		r := httptest.NewRequest("GET", "/blacklist/", data)
		w := httptest.NewRecorder()
		server.BlackList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := `{"IPList":[],"Message":{"Text":"YES","Code":0}}`
		require.Equal(t, respExp, string(respBody))
	})
	t.Run("RemoveIPInBlackList", func(t *testing.T) {//nolint: dupl, nolintlint
		t.Parallel()
		newData := storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}

		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToList(context.Background(), "blacklist", newData)
		require.NoError(t, err)
		controlDataSl, err := server.app.GetAllIPInList(context.Background(), "blacklist")
		require.NoError(t, err)
		flag := false
		for _, curControlData := range controlDataSl {
			if curControlData.IP == "192.168.16.0" && curControlData.Mask == 8 {
				flag = true

				break
			}
		}
		require.Equal(t, flag, true)

		r := httptest.NewRequest("DELETE", "/blacklist/", data)
		w := httptest.NewRecorder()
		server.BlackList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := correctOutputJSONAnswer
		require.Equal(t, respExp, string(respBody))

		controlDataSl, err = server.app.GetAllIPInList(context.Background(), "blacklist")
		require.NoError(t, err)
		flag = false
		for _, curControlData := range controlDataSl {
			if curControlData.IP == "192.168.16.0" && curControlData.Mask == 8 {
				flag = true

				break
			}
		}
		require.Equal(t, flag, false)
	})
	t.Run("GetAllIPInBlackList", func(t *testing.T) {//nolint: dupl, nolintlint
		t.Parallel()
		data := bytes.NewBufferString(`{
			"IP":"ALL",
			"Mask":0
		
		}`)
		newData := storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		server := createServer(t)

		_, err := server.app.AddIPToList(context.Background(), "blacklist", newData)
		require.NoError(t, err)
		newData = storageData.StorageIPData{
			IP:   "172.92.24.0",
			Mask: 24,
		}
		_, err = server.app.AddIPToList(context.Background(), "blacklist", newData)
		require.NoError(t, err)

		r := httptest.NewRequest("GET", "/blacklist/", data)
		w := httptest.NewRecorder()
		server.BlackList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := `{"IPList":[{"IP":"192.168.16.0","Mask":8,"ID":0},` +
			`{"IP":"172.92.24.0","Mask":24,"ID":1}],"Message":{"Text":"OK!","Code":0}}`
		require.Equal(t, respExp, string(respBody))
	})
}

func TestAuthorizationRequest(t *testing.T) {
	t.Run("AuthorizationRequest", func(t *testing.T) {//nolint: dupl, nolintlint
		data := bytes.NewBufferString(`{
			"Login":"user0",
			"Password":"CharlyDonTSerf",
			"IP":"192.168.16.56"
		}`)
		server := createServer(t)

		r := httptest.NewRequest("GET", "/request/", data)
		w := httptest.NewRecorder()
		server.AuthorizationRequest(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := `{"Message":"clear check","Ok":true}`
		require.Equal(t, respExp, string(respBody))
	})
}

func TestClearBucketByLogin(t *testing.T) {
	t.Run("ClearBucketByLogin", func(t *testing.T) {//nolint: dupl, nolintlint
		data := bytes.NewBufferString(`{
			"Tag":"user0"
		}`)
		server := createServer(t)

		r := httptest.NewRequest("DELETE", "/clearbucketbylogin/", data)
		w := httptest.NewRecorder()
		server.ClearBucketByLogin(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := correctOutputJSONAnswer
		require.Equal(t, respExp, string(respBody))
	})
}

func TestClearBucketByIP(t *testing.T) {
	t.Run("ClearBucketByIP", func(t *testing.T) {//nolint: dupl, nolintlint
		data := bytes.NewBufferString(`{
			"Tag":"192.168.16.56"
		}`)
		server := createServer(t)

		r := httptest.NewRequest("DELETE", "/clearbucketbyip/", data)
		w := httptest.NewRecorder()
		server.ClearBucketByIP(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		respExp := correctOutputJSONAnswer
		require.Equal(t, respExp, string(respBody))
	})
}

func createServer(t *testing.T) *Server {
	t.Helper()
	t.Helper()
	logger, _ := logger.New("debug")
	config := ConfigTest{}
	storage := storageSQLMock.New()
	ctxStor, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
	defer cancel()
	err := storage.Init(ctxStor, logger, &config)
	require.NoError(t, err)
	redis := RedisStorage.New()
	err = redis.InitAsMock(ctxStor, logger)
	require.NoError(t, err)
	antibruteforce := app.New(logger, storage, redis, &config)

	server := NewServer(logger, antibruteforce, &config)

	return server
}
