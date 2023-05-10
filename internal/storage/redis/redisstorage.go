package redisclient

import (
	"context"
	"strconv"

	redisMock "github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
	storageData "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
)

type RedisStorage struct {
	rdb        *redis.Client
	mockServer *redisMock.Miniredis
}

func New() *RedisStorage {
	return &RedisStorage{}
}

func (rs *RedisStorage) Init(ctx context.Context, logger storageData.Logger, config storageData.Config) error {
	rs.rdb = redis.NewClient(&redis.Options{
		Addr:     config.GetRedisAddress() + ":" + config.GetRedisPort(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := rs.rdb.Ping(ctx).Result()
	if err != nil {
		logger.Error("Redis DB ping error: " + err.Error())

		return err
	}
	rs.rdb.FlushDB(ctx)

	return nil
}

func (rs *RedisStorage) InitAsMock(ctx context.Context, logger storageData.Logger) error {
	var err error
	rs.mockServer, err = redisMock.Run()
	if err != nil {
		logger.Error("Redis DB mock init error: " + err.Error())

		return err
	}
	rs.rdb = redis.NewClient(&redis.Options{
		Addr:     rs.mockServer.Addr(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err = rs.rdb.Ping(ctx).Result()
	if err != nil {
		logger.Error("Redis DB mock ping error: " + err.Error())

		return err
	}
	rs.rdb.FlushDB(ctx)

	return nil
}

func (rs *RedisStorage) Close(ctx context.Context, logger storageData.Logger) error {
	err := rs.FlushStorage(ctx, logger)
	if err != nil {
		logger.Error("Redis DB flush error on close: " + err.Error())

		return err
	}
	if rs.mockServer != nil {
		rs.mockServer.Close()
	}
	err = rs.rdb.Close()
	if err != nil {
		logger.Error("Redis DB error on close: " + err.Error())

		return err
	}

	return nil
}

func (rs *RedisStorage) IncrementAndGetBucketValue(ctx context.Context, logger storageData.Logger, key string) (int64, error) {
	result, err := rs.rdb.Incr(ctx, key).Result()
	if err != nil {
		logger.Error("Redis DB IncrementAndGetBucketValue error: " + err.Error())

		return 0, err
	}

	return result, nil
}

func (rs *RedisStorage) SetBucketValue(ctx context.Context, logger storageData.Logger, key string, value int) error {
	strValue := strconv.Itoa(value)
	err := rs.rdb.Set(ctx, key, strValue, 0).Err()
	if err != nil {
		logger.Error("Redis DB SetBucketValue error: " + err.Error())

		return err
	}

	return nil
}

func (rs *RedisStorage) FlushStorage(ctx context.Context, _ storageData.Logger) error {
	rs.rdb.FlushDB(ctx)

	return nil
}
