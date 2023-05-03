package sql

import (
	"context"
	"database/sql"
	"errors"
	//"time"
	//"fmt"

	_ "github.com/go-sql-driver/mysql" // for driver
	helpers "github.com/skolzkyi/antibruteforce/helpers"
	storageData  "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

type storageIPData struct {
	IP                    string
	ID                    int
}

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
		dsn := helpers.StringBuild(config.GetDBUser(), ":", config.GetDBPassword(), "@tcp(",config.GetDBAddress(),":",config.GetDBPort(),")/", config.GetDBName(), "?parseTime=true") //nolint:lll
		// fmt.Println("dsn: ", dsn)
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

// WHITELIST

func(s *Storage)AddIPToWhiteList(ctx context.Context, logger storageData.Logger, IP string)(int, error) {
	stmt := "INSERT INTO whitelist(IP) VALUES (?)"                          
	res, err := s.DB.ExecContext(ctx, stmt, IP) 
	if err != nil {
		logger.Error("SQL DB exec stmt AddIPToWhiteList error: " + err.Error() + " stmt: " + stmt)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error("SQL AddIPToWhiteList get new id error: " + err.Error() + " stmt: " + stmt)
		return 0, err
	}

	return int(id), nil
}

func (s *Storage) RemoveIPInWhiteList(ctx context.Context, logger storageData.Logger, IP string) error {
	stmt := "DELETE from whitelist WHERE IP=?"

	_, err := s.DB.ExecContext(ctx, stmt, IP)
	if err != nil {
		logger.Error("SQL DB exec stmt RemoveIPInWhiteList error: " + err.Error() + " stmt: " + stmt)
		return err
	}
	return nil
}

func (s *Storage) IsIPInWhiteList(ctx context.Context,logger storageData.Logger, IP string) (bool,error) {
	stmt := "SELECT id, IP FROM whitelist WHERE IP = ?" 

	row := s.DB.QueryRowContext(ctx, stmt, IP)

	storagedIP:= &storageIPData{}

	err := row.Scan(&storagedIP.ID, &storagedIP.IP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		logger.Error("SQL row scan IsIPInWhiteList error: " + err.Error() + " stmt: " + stmt)
		return false, err
	}
	return true,nil
}

func (s *Storage) GetAllIPInWhiteList(ctx context.Context,logger storageData.Logger) ([]storageIPData,error) {
	resIP := make([]storageIPData, 0)
	
	stmt := "SELECT id, IP  FROM whitelist" 

	rows, err := s.DB.QueryContext(ctx, stmt)
	if err != nil {
		logger.Error("SQL GetAllIPInWhiteList DB query error: " + err.Error() + " stmt: " + stmt)
		return nil, err
	}

	defer rows.Close()

	storagedIP:= storageIPData{}

	for rows.Next() {
		err = rows.Scan(&storagedIP.ID, &storagedIP.IP)
		if err != nil {
			logger.Error("SQL rows scan GetAllIPInWhiteList error")
			return nil, err
		}

		resIP = append(resIP, storagedIP)
		storagedIP = storageIPData{}
	}

	if err = rows.Err(); err != nil {
		logger.Error("SQL GetAllIPInWhiteList  rows error: " + err.Error())
		return nil, err
	}

	return resIP, nil
}

// BLACKLIST

func(s *Storage)AddIPToBlackList(ctx context.Context, logger storageData.Logger, IP string)(int, error) {
	stmt := "INSERT INTO blacklist(IP) VALUES (?)"                          
	res, err := s.DB.ExecContext(ctx, stmt, IP) 
	if err != nil {
		logger.Error("SQL DB exec stmt AddIPToBlackList error: " + err.Error() + " stmt: " + stmt)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error("SQL AddIPToBlackList get new id error: " + err.Error() + " stmt: " + stmt)
		return 0, err
	}

	return int(id), nil
}

func (s *Storage) RemoveIPInBlackList(ctx context.Context, logger storageData.Logger, IP string) error {
	stmt := "DELETE from blacklist WHERE IP=?"

	_, err := s.DB.ExecContext(ctx, stmt, IP)
	if err != nil {
		logger.Error("SQL DB exec stmt RemoveIPInBlackList error: " + err.Error() + " stmt: " + stmt)
		return err
	}
	return nil
}

func (s *Storage) IsIPInBlackList(ctx context.Context,logger storageData.Logger, IP string) (bool,error) {
	stmt := "SELECT id, IP FROM blacklist WHERE IP = ?" 

	row := s.DB.QueryRowContext(ctx, stmt, IP)

	storagedIP:= &storageIPData{}

	err := row.Scan(&storagedIP.ID, &storagedIP.IP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		logger.Error("SQL row scan IsIPInBlackList error: " + err.Error() + " stmt: " + stmt)
		return false, err
	}
	return true,nil
}

func (s *Storage) GetAllIPInBlackList(ctx context.Context,logger storageData.Logger) ([]storageIPData,error) {
	resIP := make([]storageIPData, 0)
	
	stmt := "SELECT id, IP  FROM blacklist" 

	rows, err := s.DB.QueryContext(ctx, stmt)
	if err != nil {
		logger.Error("SQL GetAllIPInBlackList DB query error: " + err.Error() + " stmt: " + stmt)
		return nil, err
	}

	defer rows.Close()

	storagedIP:= storageIPData{}

	for rows.Next() {
		err = rows.Scan(&storagedIP.ID, &storagedIP.IP)
		if err != nil {
			logger.Error("SQL rows scan GetAllIPInBlackList error")
			return nil, err
		}

		resIP = append(resIP, storagedIP)
		storagedIP = storageIPData{}
	}

	if err = rows.Err(); err != nil {
		logger.Error("SQL GetAllIPInBlackList  rows error: " + err.Error())
		return nil, err
	}

	return resIP, nil
}

