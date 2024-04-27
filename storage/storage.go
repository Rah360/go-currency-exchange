package storage

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

// storage repository
type CurrencyRateStorage interface {
	Update(currency string, rate string) error
	Get() (map[string]string, error)
	Publish(rates map[string]string) error
	Subscribe() (<-chan string, error)
}

type RedisRateStore struct {
	client *redis.Client
}

func NewRedisRateStore(addr string) *RedisRateStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &RedisRateStore{
		client: rdb,
	}
}

func (store *RedisRateStore) Update(currency string, rate string) error {
	_, err := store.client.HSet("exchangeRates", currency, rate).Result()
	return err
}

func (store *RedisRateStore) Get() (map[string]string, error) {
	rates, err := store.client.HGetAll("exchangeRates").Result()
	if err != nil {
		return nil, err
	}
	return rates, nil
}

func (store *RedisRateStore) Publish(rates map[string]string) error {
	ratesJson, err := json.Marshal(rates)
	if err != nil {
		return err
	}
	return store.client.Publish("rateUpdates", ratesJson).Err()
}

func (store *RedisRateStore) Subscribe() (<-chan string, error) {
	pubsub := store.client.Subscribe("rateUpdates")
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer pubsub.Close()

		for msg := range pubsub.Channel() {
			ch <- msg.Payload
		}
	}()

	return ch, nil
}
