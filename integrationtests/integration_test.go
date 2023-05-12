//go:build integration
// +build integration

package integrationtests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"testing"

	"github.com/skolzkyi/antibruteforce/internal/logger"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql" // for driver
	redis "github.com/redis/go-redis/v9"
	helpers "github.com/skolzkyi/antibruteforce/helpers"
)

var (
	configFilePath string
	mySQL_DB       *sql.DB
	rdb            *redis.Client
	config         Config
	log            *logger.LogWrap
)

type AuthorizationRequestAnswer struct {
	Message string
	Ok      bool
}

type outputJSON struct {
	Text string
	Code int
}

type IPListAnswer struct {
	IPList  []storageData.StorageIPData
	Message outputJSON
}

type InputTag struct {
	Tag string
}

func init() {
	flag.StringVar(&configFilePath, "config", "./configs/dc/", "Path to config.env")
}

func TestMain(m *testing.M) {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	config = NewConfig()
	err := config.Init(configFilePath)
	if err != nil {
		fmt.Println(err)
	}

	log, err = logger.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("Integration tests down with error")
			os.Exit(1) //nolint:gocritic
		default:
			mySQL_DB, err = InitAndConnectDB(ctx, log, &config)
			if err != nil {
				log.Error("SQL InitAndConnectDB error: " + err.Error())
				cancel()
			}
			rdb, err = InitAndConnectRedis(ctx, log, &config)

			log.Info("Integration tests up")
			exitCode := m.Run()
			log.Info("exitCode:" + strconv.Itoa(exitCode))
			for{} //debug
			err = cleanDatabaseAndRedis(ctx)
			if err != nil {
				cancel()
			}
			err = closeDatabaseAndRedis(ctx)
			if err != nil {
				cancel()
			}
			log.Info("Integration tests down succesful")
			os.Exit(exitCode) //nolint:gocritic
		}
	}
}

func TestAddToWhiteList(t *testing.T) {
	t.Run("AddToWhiteList_Positive", func(t *testing.T) {
		url := helpers.StringBuild("http://", config.GetServerURL(), "/whitelist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "OK!")

		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()

		stmt := `SELECT IP,mask FROM whitelist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "192.168.0.0")
		require.Equal(t, mask, 24)

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
	t.Run("AddToWhiteList_NegativeListCrossCheck", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO blacklist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/whitelist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "IPData already exists in blacklist")

		stmt = `SELECT IP,mask FROM whitelist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.Truef(t, errors.Is(err, sql.ErrNoRows), "actual error %q", err)

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestRemoveFromWhiteList(t *testing.T) {
	t.Run("RemoveFromWhiteList_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO whitelist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `SELECT IP,mask FROM whitelist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "192.168.0.0")
		require.Equal(t, mask, 24)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/whitelist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)

		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "OK!")

		stmt = `SELECT IP,mask FROM whitelist WHERE IP = "192.168.0.0" AND mask=24`
		row = mySQL_DB.QueryRowContext(ctx, stmt)

		err = row.Scan(&IP, &mask)
		require.Truef(t, errors.Is(err, sql.ErrNoRows), "actual error %q", err)

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
	t.Run("RemoveFromWhiteList_NegativeNotInBase", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		url := helpers.StringBuild("http://", config.GetServerURL(), "/whitelist/")

		jsonStr := []byte(`{"IP":"192.68.0.0","Mask":24}`)

		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, storageData.ErrNoRecord.Error())
		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestIsIPInWhiteList(t *testing.T) {
	t.Run("IsIPInWhiteList_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO whitelist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `SELECT IP,mask FROM whitelist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "192.168.0.0")
		require.Equal(t, mask, 24)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/whitelist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := IPListAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Message.Text, "YES")

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
	t.Run("IsIPInWhiteList_NegativeNotInBase", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		url := helpers.StringBuild("http://", config.GetServerURL(), "/whitelist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := IPListAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Message.Text, "NO")
		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestGetAllIPInWhiteList(t *testing.T) {
	t.Run("GetAllIPInWhiteList_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO whitelist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `INSERT INTO whitelist(IP,mask) VALUES ("172.92.16.0",22)`
		_, err = mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/whitelist/")

		jsonStr := []byte(`{"IP":"ALL","Mask":0}`)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := IPListAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)

		result := make([]string, 0)
		for _, curIPSubNet := range answer.IPList {
			result = append(result, helpers.StringBuild(curIPSubNet.IP, "/", strconv.Itoa(curIPSubNet.Mask)))
		}
		require.Equal(t, len(result), 2)
		require.Equal(t, result[0], "192.168.0.0/24")
		require.Equal(t, result[1], "172.92.16.0/22")

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestAddToBlackList(t *testing.T) {
	t.Run("AddToBlackList_Positive", func(t *testing.T) {
		url := helpers.StringBuild("http://", config.GetServerURL(), "/blacklist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "OK!")

		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()

		stmt := `SELECT IP,mask FROM blacklist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "192.168.0.0")
		require.Equal(t, mask, 24)

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
	t.Run("AddToBlackList_NegativeListCrossCheck", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO whitelist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/blacklist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "IPData already exists in whitelist")

		stmt = `SELECT IP,mask FROM blacklist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.Truef(t, errors.Is(err, sql.ErrNoRows), "actual error %q", err)

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestRemoveFromBlackList(t *testing.T) {
	t.Run("RemoveFromBlackList_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO blacklist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `SELECT IP,mask FROM blacklist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "192.168.0.0")
		require.Equal(t, mask, 24)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/blacklist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)

		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "OK!")

		stmt = `SELECT IP,mask FROM blacklist WHERE IP = "192.168.0.0" AND mask=24`
		row = mySQL_DB.QueryRowContext(ctx, stmt)

		err = row.Scan(&IP, &mask)
		require.Truef(t, errors.Is(err, sql.ErrNoRows), "actual error %q", err)

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
	t.Run("RemoveFromBlackList_NegativeNotInBase", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		url := helpers.StringBuild("http://", config.GetServerURL(), "/blacklist/")

		jsonStr := []byte(`{"IP":"192.68.0.0","Mask":24}`)

		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, storageData.ErrNoRecord.Error())
		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestIsIPInBlackList(t *testing.T) {
	t.Run("IsIPInBlackList_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO blacklist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `SELECT IP,mask FROM blacklist WHERE IP = "192.168.0.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "192.168.0.0")
		require.Equal(t, mask, 24)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/blacklist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := IPListAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Message.Text, "YES")

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
	t.Run("IsIPInBlackList_NegativeNotInBase", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		url := helpers.StringBuild("http://", config.GetServerURL(), "/blacklist/")

		jsonStr := []byte(`{"IP":"192.168.0.0","Mask":24}`)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := IPListAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Message.Text, "NO")
		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestGetAllIPInBlackList(t *testing.T) {
	t.Run("GetAllIPInBlackList_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO blacklist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `INSERT INTO blacklist(IP,mask) VALUES ("172.92.16.0",22)`
		_, err = mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/blacklist/")

		jsonStr := []byte(`{"IP":"ALL","Mask":0}`)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := IPListAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)

		result := make([]string, 0)
		for _, curIPSubNet := range answer.IPList {
			result = append(result, helpers.StringBuild(curIPSubNet.IP, "/", strconv.Itoa(curIPSubNet.Mask)))
		}
		require.Equal(t, len(result), 2)
		require.Equal(t, result[0], "192.168.0.0/24")
		require.Equal(t, result[1], "172.92.16.0/22")

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestClearBucketByLogin(t *testing.T) {
	t.Run("ClearBucketByLogin_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		err := rdb.Set(ctx, "l_user0", "10", 0).Err()
		require.NoError(t, err)
		val, err := rdb.Get(ctx, "l_user0").Result()
		require.NoError(t, err)
		require.Equal(t, val, "10")
		url := helpers.StringBuild("http://", config.GetServerURL(), "/clearbucketbylogin/")

		jsonStr := []byte(`{"Tag":"user0"}`)
		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "OK!")

		val, err = rdb.Get(ctx, "l_user0").Result()
		require.NoError(t, err)
		require.Equal(t, val, "0")

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestClearBucketByIP(t *testing.T) {
	t.Run("ClearBucketByIP_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		err := rdb.Set(ctx, "ip_192.168.1.5", "10", 0).Err()
		require.NoError(t, err)
		val, err := rdb.Get(ctx, "ip_192.168.1.5").Result()
		require.NoError(t, err)
		require.Equal(t, val, "10")
		url := helpers.StringBuild("http://", config.GetServerURL(), "/clearbucketbyip/")

		jsonStr := []byte(`{"Tag":"192.168.1.5"}`)
		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := outputJSON{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Text, "OK!")

		val, err = rdb.Get(ctx, "ip_192.168.1.5").Result()
		require.NoError(t, err)
		require.Equal(t, val, "0")

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func TestAuthorizationRequest(t *testing.T) {
	t.Run("AuthorizationRequestSimple_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()

		url := helpers.StringBuild("http://", config.GetServerURL(), "/request/")

		jsonStr := []byte(`{
			"Login":"user0",
			"Password":"CharlyDonTSerf",
			"IP":"192.168.16.56"
		}`)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := AuthorizationRequestAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Ok, true)
		require.Equal(t, answer.Message, "clear check")

		val, err := rdb.Get(ctx, "l_user0").Result()
		require.NoError(t, err)
		require.Equal(t, val, "1")

		val, err = rdb.Get(ctx, "p_CharlyDonTSerf").Result()
		require.NoError(t, err)
		require.Equal(t, val, "1")

		val, err = rdb.Get(ctx, "ip_192.168.16.56").Result()
		require.NoError(t, err)
		require.Equal(t, val, "1")

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
	t.Run("AuthorizationRequestComplexSynthetic_Positive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.GetDBTimeOut())
		defer cancel()
		stmt := `INSERT INTO whitelist(IP,mask) VALUES ("172.92.16.0",24)`
		_, err := mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `SELECT IP,mask FROM whitelist WHERE IP = "172.92.16.0" AND mask=24`
		row := mySQL_DB.QueryRowContext(ctx, stmt)

		var IP string
		var mask int

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "172.92.16.0")
		require.Equal(t, mask, 24)

		stmt = `INSERT INTO blacklist(IP,mask) VALUES ("192.168.0.0",24)`
		_, err = mySQL_DB.ExecContext(ctx, stmt)
		require.NoError(t, err)

		stmt = `SELECT IP,mask FROM blacklist WHERE IP = "192.168.0.0" AND mask=24`
		row = mySQL_DB.QueryRowContext(ctx, stmt)

		err = row.Scan(&IP, &mask)
		require.NoError(t, err)

		require.Equal(t, IP, "192.168.0.0")
		require.Equal(t, mask, 24)

		url := helpers.StringBuild("http://", config.GetServerURL(), "/request/")

		jsonStr := []byte(`{
			"Login":"user0",
			"Password":"CharlyDonTSerf",
			"IP":"172.92.16.3"
		}`)

		client := &http.Client{}

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer := AuthorizationRequestAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Ok, true)
		require.Equal(t, answer.Message, "IP in whitelist")
		resp.Body.Close()

		jsonStr = []byte(`{
			"Login":"user0",
			"Password":"CharlyDonTSerf",
			"IP":"192.168.0.15"
		}`)

		req, err = http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		require.NoError(t, err)

		respBody, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer = AuthorizationRequestAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Ok, false)
		require.Equal(t, answer.Message, "IP in blacklist")
		resp.Body.Close()

		loginLimit := config.GetLimitFactorLogin()
		count := loginLimit / 3
		remDiv := loginLimit % 3
		complexData := make([][]byte, 3)
		complexData[0] = []byte(`{
			"Login":"user0",
			"Password":"CharlyDonTSerf",
			"IP":"10.0.0.4"
		}`)
		complexData[1] = []byte(`{
			"Login":"user0",
			"Password":"Freedom",
			"IP":"124.17.2.8"
		}`)
		complexData[2] = []byte(`{
			"Login":"user0",
			"Password":"FooBar",
			"IP":"92.88.6.10"
		}`)
		var wg sync.WaitGroup
		wg.Add(3)
		for i := 0; i < 3; i++ {
			i := i
			var countMod int
			if i == 0 {
				countMod = count + remDiv
			} else {
				countMod = count
			}
			go func(count int, complexData []byte) {
				defer wg.Done()
				for i := 0; i < count; i++ {
					jsonStr := complexData

					req, err = http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
					require.NoError(t, err)
					req.Header.Set("Content-Type", "application/json")

					resp, err := client.Do(req)
					require.NoError(t, err)
					defer resp.Body.Close()

					respBody, err := io.ReadAll(resp.Body)
					require.NoError(t, err)

					answer := AuthorizationRequestAnswer{}
					err = json.Unmarshal(respBody, &answer)
					require.NoError(t, err)
					require.Equal(t, answer.Ok, true)
					require.Equal(t, answer.Message, "clear check")
				}
			}(countMod, complexData[i])
		}
		wg.Wait()

		jsonStr = complexData[0]

		req, err = http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		require.NoError(t, err)

		respBody, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer = AuthorizationRequestAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Ok, false)
		require.Equal(t, answer.Message, "rate limit by login")
		resp.Body.Close()

		url = helpers.StringBuild("http://", config.GetServerURL(), "/clearbucketbylogin/")

		jsonStr = []byte(`{"Tag":"user0"}`)
		req, err = http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		require.NoError(t, err)

		respBody, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		answerCBL := outputJSON{}
		err = json.Unmarshal(respBody, &answerCBL)
		require.NoError(t, err)
		require.Equal(t, answerCBL.Text, "OK!")
		resp.Body.Close()

		url = helpers.StringBuild("http://", config.GetServerURL(), "/request/")

		jsonStr = complexData[0]

		req, err = http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		require.NoError(t, err)

		respBody, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		answer = AuthorizationRequestAnswer{}
		err = json.Unmarshal(respBody, &answer)
		require.NoError(t, err)
		require.Equal(t, answer.Ok, true)
		require.Equal(t, answer.Message, "clear check")
		resp.Body.Close()

		err = cleanDatabaseAndRedis(ctx)
		require.NoError(t, err)
	})
}

func InitAndConnectRedis(ctx context.Context, logger storageData.Logger, config storageData.Config) (*redis.Client, error) {
	select {
	case <-ctx.Done():
		return nil, storageData.ErrStorageTimeout
	default:
		defer recover()
		var err error
		rdb = redis.NewClient(&redis.Options{
			Addr:     config.GetRedisAddress() + ":" + config.GetRedisPort(),
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		_, err = rdb.Ping(ctx).Result()
		if err != nil {
			logger.Error("Redis DB ping error: " + err.Error())
			return nil, err
		}
		rdb.FlushDB(ctx)
		return rdb, nil
	}
}

func InitAndConnectDB(ctx context.Context, logger storageData.Logger, config storageData.Config) (*sql.DB, error) {
	select {
	case <-ctx.Done():
		return nil, storageData.ErrStorageTimeout
	default:
		defer recover()
		var err error
		dsn := helpers.StringBuild(config.GetDBUser(), ":", config.GetDBPassword(), "@tcp(", config.GetDBAddress(), ":", config.GetDBPort(), ")/", config.GetDBName(), "?parseTime=true") //nolint:lll

		mySQL_DBinn, err := sql.Open("mysql", dsn)
		if err != nil {
			logger.Error("SQL open error: " + err.Error())
			return nil, err
		}

		mySQL_DBinn.SetConnMaxLifetime(config.GetDBConnMaxLifetime())
		mySQL_DBinn.SetMaxOpenConns(config.GetDBMaxOpenConns())
		mySQL_DBinn.SetMaxIdleConns(config.GetDBMaxIdleConns())

		err = mySQL_DBinn.PingContext(ctx)
		if err != nil {
			logger.Error("SQL DB ping error: " + err.Error())
			return nil, err
		}

		return mySQL_DBinn, nil
	}
}

func cleanDatabaseAndRedis(ctx context.Context) error {
	rdb.FlushDB(ctx)

	stmt := "TRUNCATE TABLE OTUSAntibruteforce.whitelist"

	_, err := mySQL_DB.ExecContext(ctx, stmt)
	if err != nil {
		return err
	}

	stmt = "TRUNCATE TABLE OTUSAntibruteforce.blacklist"

	_, err = mySQL_DB.ExecContext(ctx, stmt)

	return err
}

func closeDatabaseAndRedis(ctx context.Context) error {
	err := rdb.Close()
	if err != nil {
		return err
	}

	err = mySQL_DB.Close()

	return err
}
