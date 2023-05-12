package sql

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql" // for driver
	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

type Storage struct {
	DB *sql.DB
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Init(ctx context.Context, logger storageData.Logger, config storageData.Config) error {
	err := s.Connect(ctx, logger, config)
	if err != nil {
		logger.Error("SQL connect error: " + err.Error())

		return err
	}
	err = s.DB.PingContext(ctx)
	if err != nil {
		logger.Error("SQL DB ping error: " + err.Error())

		return err
	}

	return err
}

func (s *Storage) Connect(ctx context.Context, logger storageData.Logger, config storageData.Config) error {
	select {
	case <-ctx.Done():

		return storageData.ErrStorageTimeout
	default:
		dsn := helpers.StringBuild(config.GetDBUser(), ":", config.GetDBPassword(), "@tcp(", config.GetDBAddress(), ":", config.GetDBPort(), ")/", config.GetDBName(), "?parseTime=true") //nolint:lll
		var err error
		s.DB, err = sql.Open("mysql", dsn)
		if err != nil {
			logger.Error("SQL open error: " + err.Error())

			return err
		}

		s.DB.SetConnMaxLifetime(config.GetDBConnMaxLifetime())
		s.DB.SetMaxOpenConns(config.GetDBMaxOpenConns())
		s.DB.SetMaxIdleConns(config.GetDBMaxIdleConns())

		return nil
	}
}

func (s *Storage) Close(ctx context.Context, logger storageData.Logger) error {
	select {
	case <-ctx.Done():

		return storageData.ErrStorageTimeout
	default:
		err := s.DB.Close()
		if err != nil {
			logger.Error("SQL DB close error: " + err.Error())

			return err
		}
	}

	return nil
}

func (s *Storage) AddIPToList(ctx context.Context, listname string, logger storageData.Logger, ipdata storageData.StorageIPData) (int, error) { //nolint: lll, nolintlint
	stmt := "INSERT INTO " + listname + "(mask,IP) VALUES (?,?)"
	res, err := s.DB.ExecContext(ctx, stmt, ipdata.Mask, ipdata.IP)
	if err != nil {
		logger.Error("SQL DB exec stmt AddIPToList error: " + err.Error() + " stmt: " + stmt)

		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error("SQL AddIPToList get new id error: " + err.Error() + " stmt: " + stmt)

		return 0, err
	}

	return int(id), nil
}

func (s *Storage) RemoveIPInList(ctx context.Context,listname string, logger storageData.Logger, ipdata storageData.StorageIPData) error { //nolint: lll, nolintlint
	stmt := "DELETE from " + listname + " WHERE IP=? AND mask=?"

	result, err := s.DB.ExecContext(ctx, stmt, ipdata.IP, ipdata.Mask)
	if err != nil {
		logger.Error("SQL DB exec stmt RemoveIPInList error: " + err.Error() + " stmt: " + stmt)

		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		logger.Error("SQL DB type RemoveIPInList error: " + err.Error() + " stmt: " + stmt)

		return err
	}

	if count == int64(0) {
		logger.Error("SQL DB exec stmt RemoveIPInList error: " + storageData.ErrNoRecord.Error() + " stmt: " + stmt)

		return storageData.ErrNoRecord
	}

	return nil
}

func (s *Storage) IsIPInList(ctx context.Context, listname string, logger storageData.Logger, ipdata storageData.StorageIPData) (bool, error) { //nolint: lll, nolintlint
	stmt := "SELECT id, IP FROM " + listname + " WHERE IP = ? AND mask=?"

	row := s.DB.QueryRowContext(ctx, stmt, ipdata.IP, ipdata.Mask)

	storagedIP := &storageData.StorageIPData{}

	err := row.Scan(&storagedIP.ID, &storagedIP.IP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		logger.Error("SQL row scan IsIPInList error: " + err.Error() + " stmt: " + stmt)

		return false, err
	}

	return true, nil
}

func (s *Storage) GetAllIPInList(ctx context.Context, listname string, logger storageData.Logger) ([]storageData.StorageIPData, error) { //nolint: lll, nolintlint
	resIP := make([]storageData.StorageIPData, 0)

	stmt := "SELECT id, mask, IP  FROM " + listname 

	rows, err := s.DB.QueryContext(ctx, stmt)
	if err != nil {
		logger.Error("SQL GetAllIPInList DB query error: " + err.Error() + " stmt: " + stmt)

		return nil, err
	}

	defer rows.Close()

	storagedIP := storageData.StorageIPData{}

	for rows.Next() {
		err = rows.Scan(&storagedIP.ID, &storagedIP.Mask, &storagedIP.IP)
		if err != nil {
			logger.Error("SQL rows scan GetAllIPInList error")

			return nil, err
		}

		resIP = append(resIP, storagedIP)
		storagedIP = storageData.StorageIPData{}
	}

	if err = rows.Err(); err != nil {
		logger.Error("SQL GetAllIPInList  rows error: " + err.Error())

		return nil, err
	}

	return resIP, nil
}