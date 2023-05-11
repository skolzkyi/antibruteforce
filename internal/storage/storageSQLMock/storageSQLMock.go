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

// WHITELIST

func (s *StorageMock) AddIPToWhiteList(ctx context.Context, logger storageData.Logger, value storageData.StorageIPData) (int, error) { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return 0, storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		s.mu.Lock()
		defer s.mu.Unlock()
		value.ID = s.idWL
		s.whitelist[tag] = value
		s.idWL++

		return value.ID, nil
	}
}

func (s *StorageMock) IsIPInWhiteList(ctx context.Context, _ storageData.Logger, value storageData.StorageIPData) (bool, error) { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return false, storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		s.mu.RLock()
		defer s.mu.RUnlock()
		var err error
		_, ok := s.whitelist[tag]

		return ok, err
	}
}

func (s *StorageMock) RemoveIPInWhiteList(ctx context.Context, _ storageData.Logger, value storageData.StorageIPData) error { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		_, ok := s.whitelist[tag]
		if !ok {
			return storageData.ErrNoRecord
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.whitelist, tag)

		return nil
	}
}

func (s *StorageMock) GetAllIPInWhiteList(ctx context.Context, _ storageData.Logger) ([]storageData.StorageIPData, error) { //nolint: lll, nolintlint
	resIPData := make([]storageData.StorageIPData, 0)
	select {
	case <-ctx.Done():

		return nil, storageData.ErrStorageTimeout
	default:
		s.mu.RLock()
		for _, curIPData := range s.whitelist {
			resIPData = append(resIPData, curIPData)
		}
		s.mu.RUnlock()
		sort.SliceStable(resIPData, func(i, j int) bool {
			return resIPData[i].ID < resIPData[j].ID
		})

		return resIPData, nil
	}
}

// BLACKLIST

func (s *StorageMock) AddIPToBlackList(ctx context.Context, logger storageData.Logger, value storageData.StorageIPData) (int, error) { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return 0, storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		s.mu.Lock()
		defer s.mu.Unlock()
		value.ID = s.idBL
		s.blacklist[tag] = value
		s.idBL++

		return value.ID, nil
	}
}

func (s *StorageMock) IsIPInBlackList(ctx context.Context, _ storageData.Logger, value storageData.StorageIPData) (bool, error) { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return false, storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		s.mu.RLock()
		defer s.mu.RUnlock()
		var err error
		_, ok := s.blacklist[tag]

		return ok, err
	}
}

func (s *StorageMock) RemoveIPInBlackList(ctx context.Context, _ storageData.Logger, value storageData.StorageIPData) error { //nolint: lll, nolintlint
	select {
	case <-ctx.Done():

		return storageData.ErrStorageTimeout
	default:
		tag := value.IP + "/" + strconv.Itoa(value.Mask)
		_, ok := s.blacklist[tag]
		if !ok {
			return storageData.ErrNoRecord
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.blacklist, tag)

		return nil
	}
}

func (s *StorageMock) GetAllIPInBlackList(ctx context.Context, _ storageData.Logger) ([]storageData.StorageIPData, error) { //nolint: lll, nolintlint
	resIPData := make([]storageData.StorageIPData, 0)
	select {
	case <-ctx.Done():

		return nil, storageData.ErrStorageTimeout
	default:
		s.mu.RLock()
		for _, curIPData := range s.blacklist {
			resIPData = append(resIPData, curIPData)
		}
		s.mu.RUnlock()
		sort.SliceStable(resIPData, func(i, j int) bool {
			return resIPData[i].ID < resIPData[j].ID
		})

		return resIPData, nil
	}
}
