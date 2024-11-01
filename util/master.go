package util

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

var luaIsMaster = redis.NewScript(`
if redis.call("set", KEYS[1], ARGV[1], "NX", "EX", ARGV[2]) 
	then return redis.status_reply("OK") 
end
if redis.call("get", KEYS[1]) == ARGV[1] 
	then return redis.call("set", KEYS[1], ARGV[1], "EX", ARGV[2]) 
end
`)

// CheckAndSetMaster 检查并设置为主进程，抢占失败则返回false
func CheckAndSetMaster(ctx context.Context, cli *redis.Client, key string, val string, t time.Duration) bool {
	_, err := luaIsMaster.Run(ctx, cli, []string{key}, val, t.Seconds()).Result()
	return err == nil
}
