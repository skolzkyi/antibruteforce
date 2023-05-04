package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	//"strconv"
	"strings"
	//"time"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)
/*
type Request struct {
	Login    string
	Password string
	IP       string
}
*/
/*
type storageIPData struct {
	IP                    string
	ID                    int
}
*/
type outputJSON struct {
	Text string
	Code int
}

type EventRawData struct {
	EventMessageTimeDelta int64
	Title                 string
	UserID                string
	Description           string
	DateStart             string
	DateStop              string
	ID                    int
}
type IPListAnswer struct {
	IPList  []storageData.StorageIPData
	Message outputJSON
}

type InputDate struct {
	Date string
}

var (
	ErrInJSONBadParse     = errors.New("error parsing input json")
	ErrOutJSONBadParse    = errors.New("error parsing output json")
	ErrUnsupportedMethod  = errors.New("unsupported method")
	ErrNoIDInIPHandler = errors.New("no ID in IP handler")
)

func apiErrHandler(err error, w *http.ResponseWriter) {
	W := *w
	newMessage := outputJSON{}
	newMessage.Text = err.Error()
	newMessage.Code = 1
	jsonstring, err := json.Marshal(newMessage)
	if err != nil {
		errMessage := helpers.StringBuild(http.StatusText(http.StatusInternalServerError), " (", err.Error(), ")")
		http.Error(W, errMessage, http.StatusInternalServerError)
	}

	_, err = W.Write(jsonstring)
	if err != nil {
		errMessage := helpers.StringBuild(http.StatusText(http.StatusInternalServerError), " (", err.Error(), ")")
		http.Error(W, errMessage, http.StatusInternalServerError)
	}
}

func (s *Server) helloWorld(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello world!"))
}

func (s *Server) AuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//ctx, cancel := context.WithTimeout(r.Context(), s.Config.GetDBTimeOut())
	//defer cancel()
	switch r.Method {
	case http.MethodGet:
		newRequest:=storageData.RequestAuth{}
		//newMessage := outputJSON{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &newRequest)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		fmt.Println("newRequest: ", newRequest)

	default:
		apiErrHandler(ErrUnsupportedMethod, &w)
		return
	}
}

func (s *Server) ClearBucketByLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//ctx, cancel := context.WithTimeout(r.Context(), s.Config.GetDBTimeOut())
	//defer cancel()
	switch r.Method {
	case http.MethodDelete:
		//newMessage := outputJSON{}

		path := strings.Trim(r.URL.Path, "/")
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			apiErrHandler(ErrNoIDInIPHandler, &w)
			return
		}

		Login:=pathParts[1]

		fmt.Println("ClearBucketByLogin Login: ", Login)

	default:
		apiErrHandler(ErrUnsupportedMethod, &w)
		return
	}
}

func (s *Server) ClearBucketByIP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//ctx, cancel := context.WithTimeout(r.Context(), s.Config.GetDBTimeOut())
	//defer cancel()
	switch r.Method {
	case http.MethodDelete:
		//newMessage := outputJSON{}

		path := strings.Trim(r.URL.Path, "/")
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			apiErrHandler(ErrNoIDInIPHandler, &w)
			return
		}

		IP:=pathParts[1]

		fmt.Println(" ClearBucketByIP IP: ", IP)

	default:
		apiErrHandler(ErrUnsupportedMethod, &w)
		return
	}
}

func (s *Server) WhiteList_REST(w http.ResponseWriter, r *http.Request) { //nolint:funlen
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), s.Config.GetDBTimeOut())
	defer cancel()
	switch r.Method {
	case http.MethodGet:
		fmt.Println("Get")
		
		IPListAnsw := IPListAnswer{}
		newData:=storageData.StorageIPData{}
		newMessage := outputJSON{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &newData)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}
		
		fmt.Println("newData: ", newData)
		if newData.IP=="ALL" {
			IPList,errInner:=s.app.GetAllIPInWhiteList(ctx)
			if errInner != nil {
				newMessage.Text = errInner.Error()
				newMessage.Code = 1
				w.Header().Add("ErrCustom",errInner.Error())
			} else {
				newMessage.Text = "OK!"
				newMessage.Code = 0
			}
			IPListAnsw.IPList = make([]storageData.StorageIPData, len(IPList))
			IPListAnsw.IPList = IPList
			IPListAnsw.Message = newMessage
			jsonstring, err := json.Marshal(IPListAnsw)
			if err != nil {
				apiErrHandler(err, &w)
				return
			}
			_, err = w.Write(jsonstring)
			if err != nil {
				apiErrHandler(err, &w)
				return
			}
			return
		}

		ok, errInner := s.app.IsIPInWhiteList(ctx, newData)
		if errInner != nil {
			newMessage.Text = errInner.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",errInner.Error())
		} else {
			if ok {
				newMessage.Text = "YES"
			}else {
				newMessage.Text = "NO"
			}
			newMessage.Code = 0
		}
		IPListAnsw.IPList = make([]storageData.StorageIPData, 0)
		IPListAnsw.Message = newMessage
		jsonstring, err := json.Marshal(IPListAnsw)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		_, err = w.Write(jsonstring)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		return

	case http.MethodPost:

		fmt.Println("Post")

		newData:=storageData.StorageIPData{}
		newMessage := outputJSON{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &newData)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		fmt.Println("newData: ", newData)

		id, errInner := s.app.AddIPToWhiteList(ctx, newData) 
		if errInner != nil {
			newMessage.Text = errInner.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",errInner.Error())
		} else {
			newMessage.Text = "OK!"
			newMessage.Code = id
		}

		jsonstring, err := json.Marshal(newMessage)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		_, err = w.Write(jsonstring)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		return

	case http.MethodDelete:

		fmt.Println("Delete")
		removeData:=storageData.StorageIPData{}
		newMessage := outputJSON{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &removeData)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		fmt.Println("removeData: ", removeData)
		
		errInner := s.app.RemoveIPInWhiteList(ctx, removeData)
		if errInner != nil {
			newMessage.Text = errInner.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",errInner.Error())
		} else {
			newMessage.Text = "OK!"
			newMessage.Code = 0
		}

		jsonstring, err := json.Marshal(newMessage)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		_, err = w.Write(jsonstring)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		return

	default:
		apiErrHandler(ErrUnsupportedMethod, &w)
		return
	}
}

func (s *Server) BlackList_REST(w http.ResponseWriter, r *http.Request) { //nolint:funlen
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), s.Config.GetDBTimeOut())
	defer cancel()
	switch r.Method {
	case http.MethodGet:
		fmt.Println("Get")
		
		IPListAnsw := IPListAnswer{}
		newData:=storageData.StorageIPData{}
		newMessage := outputJSON{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &newData)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}
		
		fmt.Println("newData: ", newData)
		if newData.IP=="ALL" {
			IPList,errInner:=s.app.GetAllIPInBlackList(ctx)
			if errInner != nil {
				newMessage.Text = errInner.Error()
				newMessage.Code = 1
				w.Header().Add("ErrCustom",errInner.Error())
			} else {
				newMessage.Text = "OK!"
				newMessage.Code = 0
			}
			IPListAnsw.IPList = make([]storageData.StorageIPData, len(IPList))
			IPListAnsw.IPList = IPList
			IPListAnsw.Message = newMessage
			jsonstring, err := json.Marshal(IPListAnsw)
			if err != nil {
				apiErrHandler(err, &w)
				return
			}
			_, err = w.Write(jsonstring)
			if err != nil {
				apiErrHandler(err, &w)
				return
			}
			return
		}

		ok, errInner := s.app.IsIPInBlackList(ctx, newData)
		if errInner != nil {
			newMessage.Text = errInner.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",errInner.Error())
		} else {
			if ok {
				newMessage.Text = "YES"
			}else {
				newMessage.Text = "NO"
			}
			newMessage.Code = 0
		}
		IPListAnsw.IPList = make([]storageData.StorageIPData, 0)
		IPListAnsw.Message = newMessage
		jsonstring, err := json.Marshal(IPListAnsw)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		_, err = w.Write(jsonstring)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		return

	case http.MethodPost:

		fmt.Println("Post")

		newData:=storageData.StorageIPData{}
		newMessage := outputJSON{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &newData)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		fmt.Println("newData: ", newData)

		id, errInner := s.app.AddIPToBlackList(ctx, newData) 
		if errInner != nil {
			newMessage.Text = errInner.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",errInner.Error())
		} else {
			newMessage.Text = "OK!"
			newMessage.Code = id
		}

		jsonstring, err := json.Marshal(newMessage)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		_, err = w.Write(jsonstring)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		return

	case http.MethodDelete:

		fmt.Println("Delete")
		removeData:=storageData.StorageIPData{}
		newMessage := outputJSON{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &removeData)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		fmt.Println("removeData: ", removeData)
		
		errInner := s.app.RemoveIPInBlackList(ctx, removeData)
		if errInner != nil {
			newMessage.Text = errInner.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",errInner.Error())
		} else {
			newMessage.Text = "OK!"
			newMessage.Code = 0
		}

		jsonstring, err := json.Marshal(newMessage)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		_, err = w.Write(jsonstring)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		return

	default:
		apiErrHandler(ErrUnsupportedMethod, &w)
		return
	}
}

