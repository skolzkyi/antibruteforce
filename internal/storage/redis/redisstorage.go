package redisclient

import (
    "context"
    "github.com/redis/go-redis/v9"
    //"fmt"
)

type RedisStorage type {
	rdb *Client
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
    _, err := client.Ping().Result()
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
func(rs *RedisStorage)IncrementAndGetBucketValue(ctx context.Context, logger storageData.Logger, key string)(int, error) {
    result, err := rdb.Incr(ctx, key).Int()
	if err != nil {
		logger.Error("Redis DB IncrementAndGetBucketValue error: " + err.Error())
		return 0,err
	}
    return result,nil
}
/*
func(rs *RedisStorage)GetBucketValue(ctx context.Context, logger storageData.Logger, key string, valueType string) (int,error) {

}
*/	
func(rs *RedisStorage)FlushStorage(ctx context.Context, _ storageData.Logger) error {
	rs.rdb.FlushDB(ctx)
}