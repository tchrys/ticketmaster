package redisservice

import (
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type redisService struct {
	client *redis.Client
}

var (
	instance *redisService
	once     sync.Once
)

const (
	addr     = "localhost:6379"
	password = ""
)

func GetInstance() *redisService {
	once.Do(func() {
		instance = &redisService{
			client: redis.NewClient(&redis.Options{
				Addr:     addr,
				Password: password,
				DB:       0,
			}),
		}
	})
	return instance
}

func (service *redisService) AddToBF(filterName string, elem string) error {
	_, err := service.client.Do("BF.ADD", filterName, elem).Result()
	return err
}

func (service *redisService) ExistsBF(filterName string, elem string) (int64, error) {
	rawResult, err := service.client.Do("BF.EXISTS", filterName, elem).Result()
	if err != nil {
		return 0, err
	}
	return rawResult.(int64), nil
}

func (service *redisService) ReserveBF(filterName string) error {
	_, err := service.client.Do("BF.RESERVE", filterName, "0.001", "5000000").Result()
	return err
}

func (service *redisService) SetKey(key string, value interface{}, expiration time.Duration) error {
	_, err := service.client.Set(key, value, expiration).Result()
	return err
}

func (service *redisService) GetKey(key string) (string, error) {
	return service.client.Get(key).Result()
}

func (service *redisService) AddToStream(streamName string, values map[string]interface{}) error {
	_, err := service.client.XAdd(&redis.XAddArgs{
		Stream: streamName,
		Values: values,
	}).Result()
	return err
}

func (service *redisService) ReadFromStream(streamName string, start string, count int64) ([]redis.XMessage, error) {
	return service.client.XRangeN(streamName, start, "+", count).Result()
}

func (service *redisService) StreamLength(streamName string) (int64, error) {
	return service.client.XLen(streamName).Result()
}

func (service *redisService) AddToSortSet(sortSetName string, score float64, member string) error {
	_, err := service.client.ZAdd(sortSetName, redis.Z{
		Score:  score,
		Member: member,
	}).Result()
	return err
}

func (service *redisService) SortSetRange(key string, start, stop int64) ([]string, error) {
	return service.client.ZRange(key, start, stop).Result()
}

func (service *redisService) SortSetCard(sortSetName string) (int64, error) {
	return service.client.ZCard(sortSetName).Result()
}

func (service *redisService) AddToSet(setName string, value string) (int64, error) {
	return service.client.SAdd(setName, value).Result()
}

func (service *redisService) SetIsMember(setName string, value string) (bool, error) {
	return service.client.SIsMember(setName, value).Result()
}
