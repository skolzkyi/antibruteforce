//go:build !integration
// +build !integration

package internalhttp

import (
	"bytes"
	"context"
	//"errors"
	//"fmt"
	"io"
	"net/http/httptest"
	//"os"
	//"strconv"
	"testing"
	"time"
	//"encoding/json"

	app "github.com/skolzkyi/antibruteforce/internal/app"
	storageData  "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
	storageSQLMock "github.com/skolzkyi/antibruteforce/internal/storage/storageSQLMock"
	RedisStorage "github.com/skolzkyi/antibruteforce/internal/storage/redis"
	logger "github.com/skolzkyi/antibruteforce/internal/logger"
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
	return 12
}

func (config *ConfigTest) GetLimitTimeCheck() time.Duration {
	return 60 * time.Second
}

func TestWhiteListREST(t *testing.T) {
	t.Run("AddIPToWhiteList", func(t *testing.T) {
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

		respExp := `{"Text":"OK!","Code":0}`
		require.Equal(t, respExp, string(respBody))
	})
	t.Run("IsIPInWhiteList", func(t *testing.T) {
		t.Parallel()
		newData:=storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		
		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToWhiteList(context.Background(), newData)
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
	t.Run("RemoveIPInWhiteList", func(t *testing.T) {
		t.Parallel()
		newData:=storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		
		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToWhiteList(context.Background(), newData)
		require.NoError(t, err) 
		controlDataSl,err:=server.app.GetAllIPInWhiteList(context.Background())
		require.NoError(t, err) 
		flag:=false
		for _,curControlData:=range controlDataSl {
			if curControlData.IP =="192.168.16.0" && curControlData.Mask==8 {
				flag=true
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

		respExp := `{"Text":"OK!","Code":0}`
		require.Equal(t, respExp, string(respBody))

		controlDataSl,err=server.app.GetAllIPInWhiteList(context.Background())
		require.NoError(t, err) 
		flag=false
		for _,curControlData:=range controlDataSl {
			if curControlData.IP =="192.168.16.0" && curControlData.Mask==8 {
				flag=true
				break
			}
		}
		require.Equal(t, flag, false)
	})
	t.Run("GetAllIPInWhiteList", func(t *testing.T) {
		t.Parallel()
		data := bytes.NewBufferString(`{
			"IP":"ALL",
			"Mask":0
		
		}`)
		newData:=storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		server := createServer(t)

		
		_, err := server.app.AddIPToWhiteList(context.Background(), newData)
		require.NoError(t, err) 
		newData=storageData.StorageIPData{
			IP:   "172.92.24.0",
			Mask: 24,
		}
		_, err = server.app.AddIPToWhiteList(context.Background(), newData)
		require.NoError(t, err) 

		r := httptest.NewRequest("GET", "/whitelist/", data)
		w := httptest.NewRecorder()
		server.WhiteList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err) 

		respExp := `{"IPList":[{"IP":"192.168.16.0","Mask":8,"ID":0},{"IP":"172.92.24.0","Mask":24,"ID":1}],"Message":{"Text":"OK!","Code":0}}`
		require.Equal(t, respExp, string(respBody))
	})
}

func TestBlackListREST(t *testing.T) {
	t.Run("AddIPToBlackList", func(t *testing.T) {
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

		respExp := `{"Text":"OK!","Code":0}`
		require.Equal(t, respExp, string(respBody))
	})
	t.Run("IsIPInBlackList", func(t *testing.T) {
		t.Parallel()
		newData:=storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		
		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToBlackList(context.Background(), newData)
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
	t.Run("RemoveIPInBlackList", func(t *testing.T) {
		t.Parallel()
		newData:=storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		
		data := bytes.NewBufferString(`{
			"IP":"192.168.16.0",
			"Mask":8
		
		}`)
		server := createServer(t)

		_, err := server.app.AddIPToBlackList(context.Background(), newData)
		require.NoError(t, err) 
		controlDataSl,err:=server.app.GetAllIPInBlackList(context.Background())
		require.NoError(t, err) 
		flag:=false
		for _,curControlData:=range controlDataSl {
			if curControlData.IP =="192.168.16.0" && curControlData.Mask==8 {
				flag=true
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

		respExp := `{"Text":"OK!","Code":0}`
		require.Equal(t, respExp, string(respBody))

		controlDataSl,err=server.app.GetAllIPInBlackList(context.Background())
		require.NoError(t, err) 
		flag=false
		for _,curControlData:=range controlDataSl {
			if curControlData.IP =="192.168.16.0" && curControlData.Mask==8 {
				flag=true
				break
			}
		}
		require.Equal(t, flag, false)
	})
	t.Run("GetAllIPInBlackList", func(t *testing.T) {
		t.Parallel()
		data := bytes.NewBufferString(`{
			"IP":"ALL",
			"Mask":0
		
		}`)
		newData:=storageData.StorageIPData{
			IP:   "192.168.16.0",
			Mask: 8,
		}
		server := createServer(t)

		
		_, err := server.app.AddIPToBlackList(context.Background(), newData)
		require.NoError(t, err) 
		newData=storageData.StorageIPData{
			IP:   "172.92.24.0",
			Mask: 24,
		}
		_, err = server.app.AddIPToBlackList(context.Background(), newData)
		require.NoError(t, err) 

		r := httptest.NewRequest("GET", "/blacklist/", data)
		w := httptest.NewRecorder()
		server.BlackList_REST(w, r)

		res := w.Result()
		defer res.Body.Close()
		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err) 

		respExp := `{"IPList":[{"IP":"192.168.16.0","Mask":8,"ID":0},{"IP":"172.92.24.0","Mask":24,"ID":1}],"Message":{"Text":"OK!","Code":0}}`
		require.Equal(t, respExp, string(respBody))
	})
}
/*
func TestCreateEventBadTimeBusy(t *testing.T) {
	data := bytes.NewBufferString(getTestEventData(t))
	server := createServer(t)

	r := httptest.NewRequest("POST", "/Event/", data)
	w := httptest.NewRecorder()
	server.Event_REST(w, r)

	res := w.Result()

	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	respExp := `{"Text":"OK!","Code":0}`
	require.Equal(t, respExp, string(respBody))
	res.Body.Close()

	data = bytes.NewBufferString(getTestEventData(t))
	r = httptest.NewRequest("POST", "/Event/", data)
	w = httptest.NewRecorder()
	server.Event_REST(w, r)

	res = w.Result()
	defer res.Body.Close()
	respBody, err = io.ReadAll(res.Body)
	require.NoError(t, err)

	respExp = `{"Text":"this date busy by other event","Code":1}`
	require.Equal(t, respExp, string(respBody))
}

func TestGetEvent(t *testing.T) {
	server := createServer(t)
	ctx := context.Background()
	std := time.Date(2023, 4, 20, 0, 0, 0, 1, time.Local)
	emtd := 4 * time.Hour
	id, err := server.app.CreateEvent(ctx, "testData", "USER0", "", std, std.Add(4*time.Hour), emtd)
	if err != nil {
		t.Fatal()
	}
	r := httptest.NewRequest("GET", "/Event/"+strconv.Itoa(id), nil)
	w := httptest.NewRecorder()
	server.Event_REST(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	answer := EventAnswer{}
	err = json.Unmarshal(respBody, &answer)
	require.NoError(t, err)
	require.Equal(t, answer.Events[0].ID, 0)
	require.Equal(t, answer.Events[0].Title, "testData")

	
}

func TestGetEventBadID(t *testing.T) {
	server := createServer(t)

	r := httptest.NewRequest("GET", "/Event/99", nil)
	w := httptest.NewRecorder()
	server.Event_REST(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	answer := EventAnswer{}
	err = json.Unmarshal(respBody, &answer)
	require.NoError(t, err)
	require.Equal(t, answer.Events[0].ID, 0)
	require.Equal(t, answer.Events[0].Title, "")
	require.Equal(t, answer.Message.Text, "record not searched")
	require.Equal(t, answer.Message.Code, 1)

}

func TestUpdateEvent(t *testing.T) {
	server := createServer(t)
	ctx := context.Background()
	std := time.Date(2023, 4, 20, 0, 0, 0, 1, time.Local)
	emtd := 4 * time.Hour
	id, err := server.app.CreateEvent(ctx, "testData - not Updated", "USER0", "", std, std.Add(4*time.Hour), emtd)
	if err != nil {
		t.Fatal()
	}
	data := bytes.NewBufferString(getTestEventData(t))
	r := httptest.NewRequest("PUT", "/Event/"+strconv.Itoa(id), data)
	w := httptest.NewRecorder()
	server.Event_REST(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	respExp := `{"Text":"OK!","Code":0}`
	require.Equal(t, respExp, string(respBody))
}

func TestUpdateEventBadID(t *testing.T) {
	server := createServer(t)

	data := bytes.NewBufferString(getTestEventData(t))
	r := httptest.NewRequest("PUT", "/Event/99", data)
	w := httptest.NewRecorder()
	server.Event_REST(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	respExp := `{"Text":"record not searched","Code":1}`
	require.Equal(t, respExp, string(respBody))
}

func TestDeleteEvent(t *testing.T) {
	server := createServer(t)
	ctx := context.Background()
	std := time.Date(2023, 4, 20, 0, 0, 0, 1, time.Local)
	emtd := 4 * time.Hour
	id, err := server.app.CreateEvent(ctx, "testData", "USER0", "", std, std.Add(4*time.Hour), emtd)
	if err != nil {
		t.Fatal()
	}

	r := httptest.NewRequest("DELETE", "/Event/"+strconv.Itoa(id), nil)
	w := httptest.NewRecorder()
	server.Event_REST(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	respExp := `{"Text":"OK!","Code":0}`
	require.Equal(t, respExp, string(respBody))

	_, err = server.app.GetEvent(ctx, id)
	require.Truef(t, errors.Is(err, storage.ErrNoRecord), "actual error %q", err)
}

func TestDeleteEventBadID(t *testing.T) {
	server := createServer(t)

	r := httptest.NewRequest("DELETE", "/Event/99", nil)
	w := httptest.NewRecorder()
	server.Event_REST(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	respExp := `{"Text":"record not searched","Code":1}`
	require.Equal(t, respExp, string(respBody))
}

func TestGetEventsOnDayByDay(t *testing.T) {
	server := createServer(t)
	createTestEventPool(t, server)
	r := httptest.NewRequest("GET", "/GetEventsOnDayByDay/", bytes.NewBufferString(`{"Date":"2023-04-20 17:51:00"}`))
	w := httptest.NewRecorder()
	server.GetEventsOnDayByDay(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	answer := EventAnswer{}
	err = json.Unmarshal(respBody, &answer)
	require.NoError(t, err)
	resID := make(map[int]struct{})
	for _, curEvent := range answer.Events {
		resID[curEvent.ID] = struct{}{}
	}
	_, ok := resID[0]
	require.Equal(t, ok, true)
	_, ok = resID[4]
	require.Equal(t, ok, true)
	_, ok = resID[5]
	require.Equal(t, ok, true)
	require.Equal(t, len(resID), 3)

}

func TestGetEventsOnWeekByDay(t *testing.T) {
	server := createServer(t)
	createTestEventPool(t, server)
	r := httptest.NewRequest("GET", "/GetEventsOnWeekByDay/", bytes.NewBufferString(`{"Date":"2023-04-20 17:51:00"}`))
	w := httptest.NewRecorder()
	server.GetEventsOnWeekByDay(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	answer := EventAnswer{}
	err = json.Unmarshal(respBody, &answer)
	require.NoError(t, err)
	resID := make(map[int]struct{})
	for _, curEvent := range answer.Events {
		resID[curEvent.ID] = struct{}{}
	}
	_, ok := resID[0]
	require.Equal(t, ok, true)
	_, ok = resID[1]
	require.Equal(t, ok, true)
	_, ok = resID[2]
	require.Equal(t, ok, true)
	_, ok = resID[4]
	require.Equal(t, ok, true)
	_, ok = resID[5]
	require.Equal(t, ok, true)
	require.Equal(t, len(resID), 5)
}

func TestGetEventsOnMonthByDay(t *testing.T) {
	server := createServer(t)
	createTestEventPool(t, server)
	r := httptest.NewRequest("GET", "/GetEventsOnMonthByDay/", bytes.NewBufferString(`{"Date":"2023-04-20 17:51:00"}`))
	w := httptest.NewRecorder()
	server.GetEventsOnMonthByDay(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	answer := EventAnswer{}
	err = json.Unmarshal(respBody, &answer)
	require.NoError(t, err)
	resID := make(map[int]struct{})
	for _, curEvent := range answer.Events {
		resID[curEvent.ID] = struct{}{}
	}
	_, ok := resID[0]
	require.Equal(t, ok, true)
	_, ok = resID[1]
	require.Equal(t, ok, true)
	_, ok = resID[2]
	require.Equal(t, ok, true)
	_, ok = resID[3]
	require.Equal(t, ok, true)
	_, ok = resID[4]
	require.Equal(t, ok, true)
	_, ok = resID[5]
	require.Equal(t, ok, true)
	require.Equal(t, len(resID), 6)
}

func GetListEventsNotificationByDay(t *testing.T) {
	server := createServer(t)
	createTestEventPool(t, server)
	r := httptest.NewRequest("GET", "/GetListEventsNotificationByDay/", bytes.NewBufferString(`{"Date":"2023-04-20 09:00:00"}`)) //nolint:lll,nolintlint
	w := httptest.NewRecorder()
	server.GetEventsOnMonthByDay(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	answer := EventAnswer{}
	err = json.Unmarshal(respBody, &answer)
	require.NoError(t, err)
	resID := make(map[int]struct{})
	for _, curEvent := range answer.Events {
		resID[curEvent.ID] = struct{}{}
	}
	_, ok := resID[0]
	require.Equal(t, ok, true)
	require.Equal(t, len(resID), 1)

}

func TestDeleteOldEventsByDay(t *testing.T) {
	server := createServer(t)
	createTestEventPool(t, server)

	r := httptest.NewRequest("DELETE", "/DeleteOldEventsByDay/", bytes.NewBufferString(`{"Date":"2024-04-20 09:00:00"}`)) //nolint:lll,nolintlint
	w := httptest.NewRecorder()
	server.DeleteOldEventsByDay(w, r)

	res := w.Result()
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	respExp := `{"Text":"OK!","Code":3}`
	require.Equal(t, respExp, string(respBody))

	ctx := context.Background()
	std := time.Date(2023, 4, 20, 0, 0, 0, 1, time.Local)
	events, err := server.app.GetListEventsOnMonthByDay(ctx, std)
	require.NoError(t, err)
	resID := make(map[int]struct{})
	for _, curEvent := range events {
		resID[curEvent.ID] = struct{}{}
	}
	_, ok := resID[1]
	require.Equal(t, ok, true)
	_, ok = resID[2]
	require.Equal(t, ok, true)
	_, ok = resID[3]
	require.Equal(t, ok, true)

	require.Equal(t, len(resID), 3)

}

func createTestEventPool(t *testing.T, server *Server) {
	t.Helper()
	ctx := context.Background()
	std := time.Date(2023, 4, 20, 12, 0, 0, 1, time.Local)
	emtd := 4 * time.Hour
	_, err := server.app.CreateEvent(ctx, "test0 - base event", "USER0", "", std, std.Add(4*time.Hour), emtd) 
	if err != nil {
		t.Fatal()
	}
	_, err = server.app.CreateEvent(ctx, "test1 - +5days", "USER0", "", std.Add(120*time.Hour), std.Add(124*time.Hour), emtd) //nolint:lll,nolintlint
	if err != nil {
		t.Fatal()
	}
	_, err = server.app.CreateEvent(ctx, "test2 - +6 days end date after week", "USER0", "", std.Add(144*time.Hour), std.Add(150*time.Hour), emtd) //nolint:lll,nolintlint
	if err != nil {
		t.Fatal()
	}
	_, err = server.app.CreateEvent(ctx, "test3 - +8 days - next week", "USER0", "", std.Add(192*time.Hour), std.Add(200*time.Hour), emtd) //nolint:lll,nolintlint
	if err != nil {
		t.Fatal()
	}
	_, err = server.app.CreateEvent(ctx, "test4 - start in before week and end in cur week", "USER0", "", std.Add(-48*time.Hour), std.Add(-5*time.Hour), emtd) //nolint:lll,nolintlint
	if err != nil {
		t.Fatal()
	}
	_, err = server.app.CreateEvent(ctx, "test5 - in this day", "USER0", "", std.Add(-4*time.Hour), std.Add(-3*time.Hour), emtd) //nolint:lll,nolintlint
	if err != nil {
		t.Fatal()
	}
}

func getTestEventData(t *testing.T) string {
	t.Helper()
	return `{
		"ID":                    0,
		"Title":                 "testData",
		"UserID":                "USER0",
		"Description":           "",
		"DateStart":             "2023-04-20 15:04:05",
		"DateStop":              "2023-04-22 15:04:05",
		"EventMessageTimeDelta": 10800000
	}`
}
*/
func createServer(t *testing.T) *Server {
	t.Helper()
	t.Helper()
	logger, _ := logger.New("debug")
	config:=ConfigTest{}
	var storage app.Storage
	storage = storageSQLMock.New()
	ctxStor, _ := context.WithTimeout(context.Background(), config.GetDBTimeOut())
	err := storage.Init(ctxStor, logger, &config)
	require.NoError(t, err)
	redis := RedisStorage.New()
	err = redis.InitAsMock(ctxStor, logger)
	require.NoError(t, err)
	antibruteforce := app.New(logger, storage, redis, &config)

	server := NewServer(logger, antibruteforce, &config)

	return server
}
