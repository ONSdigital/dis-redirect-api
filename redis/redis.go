package redis

import (
	"context"
	"log"

	cfg "github.com/ONSdigital/dis-redirect-api/config"
	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

type Redis struct {
	cfg.RedisConfig
}

func (r *Redis) Init(ctx context.Context) (disRedisClient *disRedis.Client, err error) {
	redisClient, redisClientErr := disRedis.NewClient(ctx, &disRedis.ClientConfig{
		Address: r.Address,
	})
	if redisClientErr != nil {
		log.Fatal(ctx, "failed to create dis-redis client", redisClientErr)
	}

	return redisClient, err
}

func (r *Redis) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	redisClient, err := r.Init(ctx)
	if err != nil {
		log.Fatal(ctx, "could not instantiate dis-redis client", err)
		return err
	}

	return redisClient.Checker(context.Background(), state)
}
