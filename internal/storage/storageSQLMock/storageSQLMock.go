package storagemock

import (
	"context"
	"sort"
	"strconv"
	"sync"

	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

type StorageMock struct {
	mu        sync.RWMutex
	whitelist map[string]storageData.StorageIPData
	blacklist map[string]storageData.StorageIPData
	idWL      int
	idBL      int
}

func New() *StorageMock {
	return &StorageMock{}
}

func (s *StorageMock) Init(_ context.Context, _ storageData.Logger, _ storageData.Config) error { //nolint: lll, nolintlint
	s.mu.Lock()
	defer s.mu.Unlock()
	s.whitelist = make(map[string]storageData.StorageIPData)
	s.blacklist = make(map[string]storageData.StorageIPData)
	s.idWL = 0
	s.idBL = 0

	return nil
}

func (s *StorageMock) Close(_ context.Context, _ storageData.Logger) error {
	return nil
}

func (s *StorageMock) AddIPToList(ctx context.Context, listname string, _ storageData.Logger, value storageData.StorageIPData) (int, error) { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return 0, storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		s.mu.Lock()
		defer s.mu.Unlock()
		switch listname {
		case storageData.WhiteListName:
			value.ID = s.idWL
			s.whitelist[tag] = value
			s.idWL++
		case storageData.BlackListName:
			value.ID = s.idBL
			s.blacklist[tag] = value
			s.idBL++
		default:
			return 0, storageData.ErrErrorBadListType
		}

		return value.ID, nil
	}
}

func (s *StorageMock) IsIPInList(ctx context.Context, listname string, _ storageData.Logger, value storageData.StorageIPData) (bool, error) { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return false, storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		s.mu.RLock()
		defer s.mu.RUnlock()
		var err error
		var ok bool
		switch listname {
		case storageData.WhiteListName:
			_, ok = s.whitelist[tag]
		case storageData.BlackListName:
			_, ok = s.blacklist[tag]
		default:
			return false, storageData.ErrErrorBadListType
		}

		return ok, err
	}
}

func (s *StorageMock) RemoveIPInList(ctx context.Context, listname string, _ storageData.Logger, value storageData.StorageIPData) error { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return storageData.ErrStorageTimeout
	default:
		var ok bool
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		switch listname {
		case storageData.WhiteListName:
			_, ok = s.whitelist[tag]
		case storageData.BlackListName:
			_, ok = s.blacklist[tag]
		default:
			return storageData.ErrErrorBadListType
		}

		if !ok {
			return storageData.ErrNoRecord
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		switch listname {
		case storageData.WhiteListName:
			delete(s.whitelist, tag)
		case storageData.BlackListName:
			delete(s.blacklist, tag)
		default:
			return storageData.ErrErrorBadListType
		}

		return nil
	}
}

func (s *StorageMock) GetAllIPInList(ctx context.Context, listname string, _ storageData.Logger) ([]storageData.StorageIPData, error) { //nolint: lll, nolintlint
	resIPData := make([]storageData.StorageIPData, 0)
	select {
	case <-ctx.Done():

		return nil, storageData.ErrStorageTimeout
	default:
		s.mu.RLock()
		switch listname {
		case storageData.WhiteListName:
			for _, curIPData := range s.whitelist {
				resIPData = append(resIPData, curIPData)
			}
		case storageData.BlackListName:
			for _, curIPData := range s.blacklist {
				resIPData = append(resIPData, curIPData)
			}
		default:
			return nil, storageData.ErrErrorBadListType
		}

		s.mu.RUnlock()
		sort.SliceStable(resIPData, func(i, j int) bool {
			return resIPData[i].ID < resIPData[j].ID
		})

		return resIPData, nil
	}
}
