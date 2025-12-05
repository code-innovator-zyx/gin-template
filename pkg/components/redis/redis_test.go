package redis

import (
	"strconv"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func getPort(addr string) int {
	parts := strings.Split(addr, ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])
	return port
}

func TestNewClient_Success(t *testing.T) {
	// Start miniredis server
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	config := Config{
		Host:     mr.Host(),
		Port:     getPort(mr.Addr()),

		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Test connection
	err = client.Ping(client.Context()).Err()
	assert.NoError(t, err)

	client.Close()
}

func TestNewClient_DefaultPoolSize(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	config := Config{
		Host:     mr.Host(),
		Port:     getPort(mr.Addr()),
		Password: "",
		DB:       0,
		PoolSize: 0, // Should default to 10
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Verify pool size defaults to 10
	assert.Equal(t, 10, client.Options().PoolSize)

	client.Close()
}

func TestNewClient_WithPassword(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	// Set password on miniredis
	mr.RequireAuth("testpassword")

	config := Config{
		Host:     mr.Host(),
		Port:     getPort(mr.Addr()),
		Password: "testpassword",
		DB:       0,
		PoolSize: 5,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	err = client.Ping(client.Context()).Err()
	assert.NoError(t, err)

	client.Close()
}

func TestNewClient_WrongPassword(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	mr.RequireAuth("correctpassword")

	config := Config{
		Host:     mr.Host(),
		Port:     getPort(mr.Addr()),
		Password: "wrongpassword",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_InvalidHost(t *testing.T) {
	config := Config{
		Host:     "invalid-host-that-does-not-exist",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewClient(config)
	assert.Error(t, err, "should fail to connect to invalid host")
	assert.Nil(t, client)
}

func TestNewClient_DifferentDB(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	config := Config{
		Host:     mr.Host(),
		Port:     getPort(mr.Addr()),
		Password: "",
		DB:       5, // Use DB 5
		PoolSize: 10,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	assert.Equal(t, 5, client.Options().DB)

	client.Close()
}

func TestNewClient_CustomPoolSize(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	config := Config{
		Host:     mr.Host(),
		Port:     getPort(mr.Addr()),
		Password: "",
		DB:       0,
		PoolSize: 25,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	assert.Equal(t, 25, client.Options().PoolSize)

	client.Close()
}
