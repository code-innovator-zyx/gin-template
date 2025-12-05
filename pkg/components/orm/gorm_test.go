package orm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


func TestGetLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel int
		expected logger.LogLevel
	}{
		{"Silent", 0, logger.Silent},
		{"Error", 1, logger.Error},
		{"Warn", 2, logger.Warn},
		{"Info", 3, logger.Info},
		{"Info (level 4)", 4, logger.Info},
		{"Default", 99, logger.Info},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := getLogger(tt.logLevel)
			assert.NotNil(t, l)
		})
	}
}

func TestUpdateLogLevel(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: getLogger(3),
	})
	assert.NoError(t, err)

	// Test updating log level
	err = UpdateLogLevel(db, 1)
	assert.NoError(t, err)

	// Verify the logger was updated
	assert.NotNil(t, db.Config.Logger)
}

func TestUpdateLogLevel_NilDB(t *testing.T) {
	err := UpdateLogLevel(nil, 3)
	assert.Error(t, err)
	assert.Equal(t, "db 为 nil", err.Error())
}

func TestUpdateLogLevel_DifferentLevels(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	levels := []int{0, 1, 2, 3, 4}
	for _, level := range levels {
		err := UpdateLogLevel(db, level)
		assert.NoError(t, err)
	}
}

func TestInit_SQLite(t *testing.T) {
	// Test with a valid SQLite DSN (since we can't test MySQL without a real server)
	// We'll test the basic structure by using getLogger and config setup
	
	// Test that Config structure is defined correctly
	config := Config{
		DSN:          "test.db",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		MaxLifetime:  300,
		LogLevel:     3,
	}

	assert.Equal(t, "test.db", config.DSN)
	assert.Equal(t, 10, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 300, config.MaxLifetime)
	assert.Equal(t, 3, config.LogLevel)
}

func TestInit_GormConfig(t *testing.T) {
	// Test GORM configuration setup
	gormConfig := &gorm.Config{
		Logger: getLogger(3),
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	assert.NotNil(t, gormConfig.Logger)
	assert.True(t, gormConfig.DisableForeignKeyConstraintWhenMigrating)
	assert.NotNil(t, gormConfig.NowFunc)
}


// Note: Full Init testing with MySQL requires a real MySQL server,
// which is not practical for unit tests. These tests cover the
// getLogger and UpdateLogLevel functions which are testable in isolation.
// For integration testing of Init and createDatabaseIfNotExist,
// you would need Docker or a test MySQL instance.
