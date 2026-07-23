package infra

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
)

type redisRESPClient struct {
	connection net.Conn
	reader     *bufio.Reader
}

func dialRedis() (redisRESPClient, error) {
	address := os.Getenv("NIMBUS_REDIS_ADDR")
	if address == "" {
		address = "127.0.0.1:27316"
	}
	connection, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return redisRESPClient{}, err
	}
	client := redisRESPClient{connection: connection, reader: bufio.NewReader(connection)}
	if password := os.Getenv("NIMBUS_REDIS_PASSWORD"); password != "" {
		if _, err = client.command("AUTH", password); err != nil {
			connection.Close()
			return redisRESPClient{}, err
		}
	}
	return client, nil
}
func (c redisRESPClient) close() { _ = c.connection.Close() }
func (c redisRESPClient) command(parts ...string) (string, error) {
	var request strings.Builder
	fmt.Fprintf(&request, "*%d\r\n", len(parts))
	for _, part := range parts {
		fmt.Fprintf(&request, "$%d\r\n%s\r\n", len(part), part)
	}
	_ = c.connection.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := c.connection.Write([]byte(request.String())); err != nil {
		return "", err
	}
	prefix, err := c.reader.ReadByte()
	if err != nil {
		return "", err
	}
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	switch prefix {
	case '+', ':':
		return line, nil
	case '-':
		return "", fmt.Errorf("Redis: %s", line)
	case '$':
		length, parseErr := strconv.Atoi(line)
		if parseErr != nil {
			return "", parseErr
		}
		if length < 0 {
			return "", nil
		}
		buffer := make([]byte, length+2)
		if _, err = io.ReadFull(c.reader, buffer); err != nil {
			return "", err
		}
		return string(buffer[:length]), nil
	default:
		return "", fmt.Errorf("未知 Redis RESP 类型 %q", prefix)
	}
}
func parseRedisInfo(raw string) map[string]string {
	info := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			info[parts[0]] = parts[1]
		}
	}
	return info
}

// RedisMonitor godoc
// @Summary Get live Redis monitor information
// @Tags Infra Redis
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/redis/get-monitor-info [get]
func (h *Handler) RedisMonitor(c *gin.Context) {
	client, err := dialRedis()
	if err != nil {
		httpx.Fail(c, 503, 503, "Redis 不可用: "+err.Error())
		return
	}
	defer client.close()
	raw, err := client.command("INFO", "ALL")
	if err != nil {
		httpx.Fail(c, 503, 503, "读取 Redis INFO 失败: "+err.Error())
		return
	}
	sizeRaw, err := client.command("DBSIZE")
	if err != nil {
		httpx.Fail(c, 503, 503, "读取 Redis DBSIZE 失败: "+err.Error())
		return
	}
	size, _ := strconv.ParseInt(sizeRaw, 10, 64)
	info := parseRedisInfo(raw)
	stats := []gin.H{}
	for key, value := range info {
		if !strings.HasPrefix(key, "cmdstat_") {
			continue
		}
		fields := map[string]int64{}
		for _, pair := range strings.Split(value, ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				fields[parts[0]], _ = strconv.ParseInt(parts[1], 10, 64)
			}
		}
		stats = append(stats, gin.H{"command": strings.TrimPrefix(key, "cmdstat_"), "calls": fields["calls"], "usec": fields["usec"]})
	}
	httpx.OK(c, gin.H{"info": info, "dbSize": size, "commandStats": stats})
}
