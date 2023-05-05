package redisclient

import (
    "context"
    "strconv"
    redis "github.com/redis/go-redis/v9"
    storageData  "github.com/skolzkyi/antibruteforce/internal/storage/storageData"
    //"fmt"
)

type RedisStorage struct {
	rdb *redis.Client
}


func New() *RedisStorage {
	return &RedisStorage{}
}

func(rs *RedisStorage)Init(ctx context.Context, logger storageData.Logger, config storageData.Config) error {
    rs.rdb = redis.NewClient(&redis.Options{
        Addr:     config.GetResisAddress()+":"+config.GetRedisPort(),
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
/*
func(rs *RedisStorage)Close(ctx context.Context, logger storageData.Logger) error {

}
*/
func(rs *RedisStorage)IncrementAndGetBucketValue(ctx context.Context, logger storageData.Logger, key string)(int64, error) {
    result, err := rs.rdb.Incr(ctx, key).Result()
	if err != nil {
		logger.Error("Redis DB IncrementAndGetBucketValue error: " + err.Error())
		return 0,err
	}

    return result,nil
}

func(rs *RedisStorage)SetBucketValue(ctx context.Context, logger storageData.Logger, key string, value int) (error) {
    strValue:=strconv.Itoa(value)
    err := rs.rdb.Set(ctx, key, strValue, 0).Err()
	if err != nil {
		logger.Error("Redis DB SetBucketValue error: " + err.Error())
		return err
	}

    return nil
}
	
func(rs *RedisStorage)FlushStorage(ctx context.Context, _ storageData.Logger) error {
	rs.rdb.FlushDB(ctx)
    return nil
}