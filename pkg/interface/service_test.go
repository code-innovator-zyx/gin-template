package _interface

import (
	"context"
	cache2 "gin-admin/pkg/components/cache"
	"gorm.io/driver/mysql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/12/2
* @Package: Service 测试用例（缓存功能）
 */

// setupTestService 创建测试 Service
func setupTestService(t *testing.T) (*Service[TestUser], *gorm.DB, cache2.ICache) {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// 使用内存缓存
	cacheInstance := cache2.NewMemoryCache()
	require.NoError(t, err)
	service := NewService[TestUser](db, cacheInstance)

	return service, db, cacheInstance
}

// TestService_FindByID_Cache 测试 FindByID 的缓存功能
func TestService_FindByID_Cache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Username: "testcase4", Email: "cache@example.com", Age: 25}
	db.Create(user)

	t.Run("第一次查询，缓存未命中", func(t *testing.T) {
		result, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "testcase4", result.Username)
	})

	t.Run("第二次查询，缓存命中", func(t *testing.T) {
		// 修改数据库中的数据
		db.Model(&TestUser{}).Where("id = ?", user.ID).Update("age", 89)

		// 从 Service 查询，应该返回缓存的数据（age=25）
		result, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 25, result.Age) // 缓存的值
	})

	t.Run("清空缓存后重新查询", func(t *testing.T) {
		err := service.ClearCache(ctx)
		assert.NoError(t, err)

		// 应该从数据库查询到更新后的值
		result, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 89, result.Age) // 数据库的新值
	})
}

// TestService_FindOne_Cache 测试 FindOne 的缓存功能
func TestService_FindOne_Cache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	user := &TestUser{Username: "findone", Email: "findone@example.com", Age: 30}
	db.Create(user)

	t.Run("缓存命中测试", func(t *testing.T) {
		// 第一次查询
		result1, err := service.FindOne(ctx, WithConditions(map[string]interface{}{
			"username": "findone",
		}))
		assert.NoError(t, err)
		assert.Equal(t, 30, result1.Age)

		// 修改数据库
		db.Model(&TestUser{}).Where("username = ?", "findone").Update("age", 88)

		// 第二次查询，应该返回缓存值
		result2, err := service.FindOne(ctx, WithConditions(map[string]interface{}{
			"username": "findone",
		}))
		assert.NoError(t, err)
		assert.Equal(t, 30, result2.Age) // 缓存值
	})
}

// TestService_List_Cache 测试 List 的缓存功能
func TestService_List_Cache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	// 创建多条数据
	users := []TestUser{
		{Username: "list1", Email: "list1@example.com", Status: 1},
		{Username: "list2", Email: "list2@example.com", Status: 1},
		{Username: "list3", Email: "list3@example.com", Status: 0},
	}
	for i := range users {
		db.Create(&users[i])
	}

	t.Run("列表缓存测试", func(t *testing.T) {
		// 第一次查询
		list1, err := service.List(ctx, WithConditions(map[string]interface{}{
			"status": 1,
		}))
		assert.NoError(t, err)
		assert.Len(t, list1, 7)

		// 添加新数据
		newUser := &TestUser{Username: "list4", Email: "list4@example.com", Status: 1}
		db.Create(newUser)

		// 第二次查询，应该返回缓存（还是2条）
		list2, err := service.List(ctx, WithConditions(map[string]interface{}{
			"status": 1,
		}))
		assert.NoError(t, err)
		assert.Len(t, list2, 7) // 缓存值
	})
}

// TestService_FindPage_Cache 测试分页查询的缓存
func TestService_FindPage_Cache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试数据
	for i := 1; i <= 10; i++ {
		user := &TestUser{
			Username: "page" + string(rune('0'+i)),
			Email:    "page" + string(rune('0'+i)) + "@example.com",
		}
		db.Create(user)
	}

	t.Run("分页缓存测试", func(t *testing.T) {
		// 第一次查询
		page1, err := service.FindPage(ctx, WithPagination(1, 3))
		assert.NoError(t, err)
		assert.Equal(t, int64(10), page1.Total)
		assert.Len(t, page1.List, 3)

		// 添加新数据
		db.Create(&TestUser{Username: "page11", Email: "page11@example.com"})

		// 第二次查询，应该返回缓存（total=10）
		page2, err := service.FindPage(ctx, WithPagination(1, 3))
		assert.NoError(t, err)
		assert.Equal(t, int64(10), page2.Total) // 缓存值
	})
}

// TestService_FindByIDs_NoCache 测试批量查询不缓存
func TestService_FindByIDs_NoCache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	users := []TestUser{
		{Username: "batch1", Email: "batch1@example.com", Age: 20},
		{Username: "batch2", Email: "batch2@example.com", Age: 21},
	}
	for i := range users {
		db.Create(&users[i])
	}

	t.Run("批量查询不缓存", func(t *testing.T) {
		ids := []uint{users[0].ID, users[1].ID}

		// 第一次查询
		result1, err := service.FindByIDs(ctx, ids)
		assert.NoError(t, err)
		assert.Len(t, result1, 2)

		// 修改数据库
		db.Model(&TestUser{}).Where("id = ?", users[0].ID).Update("age", 99)

		// 第二次查询，应该返回数据库的新值（不走缓存）
		result2, err := service.FindByIDs(ctx, ids)
		assert.NoError(t, err)
		assert.Equal(t, 99, result2[0].Age) // 新值，证明没走缓存
	})
}

// TestService_Create_InvalidateCache 测试创建后缓存失效
func TestService_Create_InvalidateCache(t *testing.T) {
	service, _, _ := setupTestService(t)
	ctx := context.Background()

	t.Run("创建后列表缓存失效", func(t *testing.T) {
		// 先查询列表并缓存
		list1, err := service.List(ctx)
		assert.NoError(t, err)
		initialCount := len(list1)

		// 创建新数据
		newUser := &TestUser{Username: "invalidate", Email: "invalidate@example.com"}
		err = service.Create(ctx, newUser)
		assert.NoError(t, err)

		// 再次查询，应该返回新数据（缓存已失效）
		list2, err := service.List(ctx)
		assert.NoError(t, err)
		assert.Equal(t, initialCount+1, len(list2))
	})
}

// TestService_Update_InvalidateCache 测试更新后缓存失效
func TestService_Update_InvalidateCache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	user := &TestUser{Username: "update", Email: "update@example.com", Age: 25}
	db.Create(user)

	t.Run("UpdateByID 后缓存失效", func(t *testing.T) {
		// 先查询并缓存
		result1, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 25, result1.Age)

		// 更新
		err = service.UpdateByID(ctx, user.ID, map[string]interface{}{
			"age": 50,
		})
		assert.NoError(t, err)

		// 再次查询，应该返回新值（缓存已失效）
		result2, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 50, result2.Age)
	})

	t.Run("Update 后缓存失效", func(t *testing.T) {
		result, _ := service.FindByID(ctx, user.ID)
		result.Age = 60

		err := service.Update(ctx, result)
		assert.NoError(t, err)

		// 验证缓存已失效
		updated, _ := service.FindByID(ctx, user.ID)
		assert.Equal(t, 60, updated.Age)
	})
}

// TestService_Delete_InvalidateCache 测试删除后缓存失效
func TestService_Delete_InvalidateCache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	user := &TestUser{Username: "delete", Email: "delete@example.com"}
	db.Create(user)

	t.Run("DeleteByID 后缓存失效", func(t *testing.T) {
		// 先查询并缓存
		_, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)

		// 查询列表并缓存
		list1, _ := service.List(ctx)
		initialCount := len(list1)

		// 删除
		err = service.DeleteByID(ctx, user.ID)
		assert.NoError(t, err)

		// 验证 FindByID 缓存已失效
		_, err = service.FindByID(ctx, user.ID)
		assert.Error(t, err)

		// 验证列表缓存已失效
		list2, _ := service.List(ctx)
		assert.Equal(t, initialCount-1, len(list2))
	})
}

// TestService_Transaction_ClearCache 测试事务后清空缓存
func TestService_Transaction_ClearCache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	user := &TestUser{Username: "tx", Email: "tx@example.com", Age: 30}
	db.Create(user)

	t.Run("事务成功后清空所有缓存", func(t *testing.T) {
		// 先查询并缓存
		result1, _ := service.FindByID(ctx, user.ID)
		assert.Equal(t, 30, result1.Age)

		// 执行事务
		err := service.Transaction(ctx, func(ctx context.Context, tx *gorm.DB, txRepo IRepo[TestUser]) error {
			return txRepo.UpdateByID(ctx, user.ID, map[string]interface{}{
				"age": 70,
			})
		})
		assert.NoError(t, err)

		// 验证缓存已清空（应该返回新值）
		result2, _ := service.FindByID(ctx, user.ID)
		assert.Equal(t, 70, result2.Age)
	})
}

// TestService_WithoutCache 测试没有缓存时的降级行为
func TestService_WithoutCache(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&TestUser{})

	// 传入 nil 缓存
	service := NewService[TestUser](db, nil)
	ctx := context.Background()

	user := &TestUser{Username: "nocache", Email: "nocache@example.com", Age: 40}
	db.Create(user)

	t.Run("无缓存时正常工作", func(t *testing.T) {
		// 应该直接查询数据库
		result1, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 40, result1.Age)

		// 更新数据库
		db.Model(&TestUser{}).Where("id = ?", user.ID).Update("age", 80)

		// 再次查询，应该返回新值（因为没缓存）
		result2, err := service.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 80, result2.Age)
	})
}

// TestService_Count_NoCache 测试统计操作不缓存
func TestService_Count_NoCache(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	db.Create(&TestUser{Username: "count1", Email: "count1@example.com"})

	t.Run("Count 不缓存", func(t *testing.T) {
		count1, _ := service.Count(ctx, nil)
		assert.Equal(t, int64(1), count1)

		// 添加新数据
		db.Create(&TestUser{Username: "count2", Email: "count2@example.com"})

		// 应该立即返回新值（不走缓存）
		count2, _ := service.Count(ctx, nil)
		assert.Equal(t, int64(2), count2)
	})
}

// TestService_CacheTTL 测试缓存过期
func TestService_CacheTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过耗时测试")
	}

	service, db, _ := setupTestService(t)
	// 设置极短的 TTL
	service.cacheTTL = 100 * time.Millisecond
	ctx := context.Background()

	user := &TestUser{Username: "ttl", Email: "ttl@example.com", Age: 25}
	db.Create(user)

	t.Run("缓存过期后重新查询", func(t *testing.T) {
		// 第一次查询并缓存
		result1, _ := service.FindByID(ctx, user.ID)
		assert.Equal(t, 25, result1.Age)

		// 修改数据库
		db.Model(&TestUser{}).Where("id = ?", user.ID).Update("age", 99)

		// 等待缓存过期
		time.Sleep(200 * time.Millisecond)

		// 再次查询，应该返回新值
		result2, _ := service.FindByID(ctx, user.ID)
		assert.Equal(t, 99, result2.Age)
	})
}
