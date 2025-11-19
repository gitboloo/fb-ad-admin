package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"backend/configs"
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {
	cfg := configs.AppConfig.Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Println("Failed to connect to Redis:", err)
		log.Println("Warning: Redis is not available, some features may not work properly")
		RedisClient = nil // 设置为nil,避免后续使用出错
		return
	}

	log.Println("Redis connected successfully")
}

func GetRedis() *redis.Client {
	return RedisClient
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// RedisHealthCheck 检查Redis连接健康状态
func RedisHealthCheck() error {
	if RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	
	return nil
}

// GetRedisInfo 获取Redis服务器信息
func GetRedisInfo() (map[string]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("redis not initialized")
	}
	
	ctx := context.Background()
	info, err := RedisClient.Info(ctx, "server", "clients", "memory", "stats").Result()
	if err != nil {
		return nil, err
	}
	
	// 解析info字符串为map
	infoMap := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				infoMap[parts[0]] = parts[1]
			}
		}
	}
	
	return infoMap, nil
}