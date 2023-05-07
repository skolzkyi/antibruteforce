package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	//"strconv"
	//"strings"
	"time"

	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
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
	w.Write([]byte("test"))
}

func (s *Server) AuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	switch r.Method {
	case http.MethodGet:
		newRequest:=storageData.RequestAuth{}

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

		answer:=AuthorizationRequestAnswer{}

		ok,message,errInner := s.app.CheckInputRequest(ctx, newRequest)
		if errInner != nil {
			answer.Message = "Inner error: "+errInner.Error()
			answer.Ok = false
			w.Header().Add("ErrCustom",errInner.Error())
		}
		answer.Message = message
		answer.Ok = ok
		jsonstring, err := json.Marshal(answer)
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

func (s *Server) ClearBucketByLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	switch r.Method {
	case http.MethodDelete:
		newMessage := outputJSON{}
		inputTag:=InputTag{}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &inputTag)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}
        
		fmt.Println("ClearBucketByLogin Login: ", inputTag.Tag)
		err = s.app.ClearBucketByLogin(ctx,inputTag.Tag)
		if err != nil {
			newMessage.Text = err.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",err.Error())
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

func (s *Server) ClearBucketByIP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	switch r.Method {
	case http.MethodDelete:
		newMessage := outputJSON{}
		inputTag:=InputTag{}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}

		err = json.Unmarshal(body, &inputTag)
		if err != nil {
			apiErrHandler(err, &w)
			return
		}
        
		fmt.Println(" ClearBucketByIP IP: ", inputTag.Tag)
		err =s.app.ClearBucketByIP(ctx,inputTag.Tag)
		if err != nil {
			newMessage.Text = err.Error()
			newMessage.Code = 1
			w.Header().Add("ErrCustom",err.Error())
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

