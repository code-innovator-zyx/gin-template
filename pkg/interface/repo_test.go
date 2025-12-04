package _interface

import (
	"context"
	"gorm.io/driver/sqlite"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/12/2
* @Package: Repo 测试用例
 */

// TestUser 测试用模型
type TestUser struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"size:50;uniqueIndex"`
	Email    string
	Age      int
	Status   int
}

func (TestUser) TableName() string {
	return "test_users"
}

func (u TestUser) GetID() uint {
	return u.ID
}

func (u TestUser) SetID(id uint) {
	u.ID = id
}

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	return db
}

// seedTestData 填充测试数据
func seedTestData(t *testing.T, db *gorm.DB) {
	users := []TestUser{
		{Username: "user1", Email: "user1@example.com", Age: 20, Status: 1},
		{Username: "user2", Email: "user2@example.com", Age: 25, Status: 1},
		{Username: "user3", Email: "user3@example.com", Age: 30, Status: 0},
		{Username: "user4", Email: "user4@example.com", Age: 35, Status: 1},
		{Username: "user5", Email: "user5@example.com", Age: 40, Status: 0},
	}

	for i := range users {
		err := db.Create(&users[i]).Error
		require.NoError(t, err)
	}
}

// TestRepo_FindByID 测试通过 ID 查询
func TestRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	//seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("查询存在的记录", func(t *testing.T) {
		user, err := repo.FindByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user1", user.Username)
	})

	t.Run("查询不存在的记录", func(t *testing.T) {
		user, err := repo.FindByID(ctx, 999)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

// TestRepo_FindByIDs 测试批量查询
func TestRepo_FindByIDs(t *testing.T) {
	db := setupTestDB(t)
	//seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("查询多个存在的记录", func(t *testing.T) {
		users, err := repo.FindByIDs(ctx, []uint{1, 2, 3})
		assert.NoError(t, err)
		assert.Len(t, users, 3)
	})

	t.Run("空ID列表", func(t *testing.T) {
		users, err := repo.FindByIDs(ctx, []uint{})
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

// TestRepo_FindOne 测试单条查询
func TestRepo_FindOne(t *testing.T) {
	db := setupTestDB(t)
	//seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("按条件查询", func(t *testing.T) {
		user, err := repo.FindOne(ctx, WithConditions(map[string]interface{}{
			"username": "user1",
		}))
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user1", user.Username)
	})

	t.Run("查询不存在的记录", func(t *testing.T) {
		user, err := repo.FindOne(ctx, WithConditions(map[string]interface{}{
			"username": "nonexistent",
		}))
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

// TestRepo_List 测试列表查询
func TestRepo_List(t *testing.T) {
	db := setupTestDB(t)
	//seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("查询所有", func(t *testing.T) {
		users, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 5)
	})

	t.Run("按条件查询", func(t *testing.T) {
		users, err := repo.List(ctx, WithConditions(map[string]interface{}{
			"status": 1,
		}))
		assert.NoError(t, err)
		assert.Len(t, users, 3)
	})

	t.Run("带排序查询", func(t *testing.T) {
		users, err := repo.List(ctx,
			WithConditions(map[string]interface{}{"status": 1}),
			WithOrderBy("age desc"),
		)
		assert.NoError(t, err)
		assert.Len(t, users, 3)
		assert.Equal(t, uint(4), users[0].ID) // age=35
		assert.Equal(t, uint(2), users[1].ID) // age=25
		assert.Equal(t, uint(1), users[2].ID) // age=20
	})
}

// TestRepo_FindPage 测试分页查询
func TestRepo_FindPage(t *testing.T) {
	db := setupTestDB(t)
	//seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("第一页", func(t *testing.T) {
		result, err := repo.FindPage(ctx,
			WithPagination(1, 2),
			WithOrderBy("id asc"),
		)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(5), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.PageSize)
		assert.Equal(t, 3, result.TotalPage)
		assert.Len(t, result.List, 2)
	})

	t.Run("最后一页", func(t *testing.T) {
		result, err := repo.FindPage(ctx,
			WithPagination(3, 2),
			WithOrderBy("id asc"),
		)
		assert.NoError(t, err)
		assert.Len(t, result.List, 1) // 最后一页只有1条
	})

	t.Run("超出范围的页码", func(t *testing.T) {
		result, err := repo.FindPage(ctx,
			WithPagination(10, 2),
		)
		assert.NoError(t, err)
		assert.Empty(t, result.List)
		assert.Equal(t, int64(5), result.Total)
	})
}

// TestRepo_Create 测试创建
func TestRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("创建成功", func(t *testing.T) {
		user := &TestUser{
			Username: "newuser",
			Email:    "new@example.com",
			Age:      25,
			Status:   1,
		}
		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})
}

// TestRepo_CreateBatch 测试批量创建
func TestRepo_CreateBatch(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("批量创建", func(t *testing.T) {
		users := []TestUser{
			{Username: "batch1", Email: "batch1@example.com", Age: 20},
			{Username: "batch2", Email: "batch2@example.com", Age: 21},
			{Username: "batch3", Email: "batch3@example.com", Age: 22},
		}
		err := repo.CreateBatch(ctx, users)
		assert.NoError(t, err)
	})

	t.Run("空列表", func(t *testing.T) {
		err := repo.CreateBatch(ctx, []TestUser{})
		assert.NoError(t, err)
	})
}

// TestRepo_Update 测试更新
func TestRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	//seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("更新成功", func(t *testing.T) {
		user, _ := repo.FindByID(ctx, 1)
		user.Age = 99
		err := repo.Update(ctx, user)
		assert.NoError(t, err)

		// 验证
		updated, _ := repo.FindByID(ctx, 1)
		assert.Equal(t, 99, updated.Age)
	})
}

// TestRepo_UpdateByID 测试按ID更新
func TestRepo_UpdateByID(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("更新成功", func(t *testing.T) {
		err := repo.UpdateByID(ctx, 1, map[string]interface{}{
			"age": 88,
		})
		assert.NoError(t, err)

		// 验证
		user, _ := repo.FindByID(ctx, 1)
		assert.Equal(t, 88, user.Age)
	})
}

// TestRepo_UpdateByCondition 测试按条件更新
func TestRepo_UpdateByCondition(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("批量更新", func(t *testing.T) {
		err := repo.UpdateByCondition(ctx,
			map[string]interface{}{"status": 1},
			map[string]interface{}{"age": 100},
		)
		assert.NoError(t, err)

		// 验证
		users, _ := repo.List(ctx, WithConditions(map[string]interface{}{
			"status": 1,
		}))
		for _, user := range users {
			assert.Equal(t, 100, user.Age)
		}
	})
}

// TestRepo_Delete 测试删除
func TestRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("删除成功", func(t *testing.T) {
		user, _ := repo.FindByID(ctx, 1)
		err := repo.Delete(ctx, user)
		assert.NoError(t, err)

		// 验证
		_, err = repo.FindByID(ctx, 1)
		assert.Error(t, err)
	})
}

// TestRepo_DeleteByID 测试按ID删除
func TestRepo_DeleteByID(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("删除成功", func(t *testing.T) {
		err := repo.DeleteByID(ctx, 2)
		assert.NoError(t, err)

		// 验证
		_, err = repo.FindByID(ctx, 2)
		assert.Error(t, err)
	})
}

// TestRepo_DeleteByIDs 测试批量删除
func TestRepo_DeleteByIDs(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("批量删除", func(t *testing.T) {
		err := repo.DeleteByIDs(ctx, []uint{1, 2, 3})
		assert.NoError(t, err)

		// 验证
		count, _ := repo.Count(ctx, nil)
		assert.Equal(t, int64(2), count)
	})
}

// TestRepo_DeleteByCondition 测试按条件删除
func TestRepo_DeleteByCondition(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("按条件删除", func(t *testing.T) {
		err := repo.DeleteByCondition(ctx, map[string]interface{}{
			"status": 0,
		})
		assert.NoError(t, err)

		// 验证
		count, _ := repo.Count(ctx, nil)
		assert.Equal(t, int64(3), count)
	})

	t.Run("空条件应该报错", func(t *testing.T) {
		err := repo.DeleteByCondition(ctx, map[string]interface{}{})
		assert.Error(t, err)
	})
}

// TestRepo_Count 测试统计
func TestRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("统计全部", func(t *testing.T) {
		count, err := repo.Count(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	t.Run("按条件统计", func(t *testing.T) {
		count, err := repo.Count(ctx, map[string]interface{}{
			"status": 1,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

// TestRepo_Exists 测试存在性检查
func TestRepo_Exists(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("存在的记录", func(t *testing.T) {
		exists, err := repo.Exists(ctx, WithConditions(map[string]interface{}{
			"username": "user1",
		}))
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("不存在的记录", func(t *testing.T) {
		exists, err := repo.Exists(ctx, WithConditions(map[string]interface{}{
			"username": "nonexistent",
		}))
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestRepo_ExistsByID 测试按ID检查存在性
func TestRepo_ExistsByID(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("存在的ID", func(t *testing.T) {
		exists, err := repo.ExistsByID(ctx, 1)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("不存在的ID", func(t *testing.T) {
		exists, err := repo.ExistsByID(ctx, 999)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestRepo_FirstOrCreate 测试查找或创建
func TestRepo_FirstOrCreate(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("记录已存在", func(t *testing.T) {
		user := &TestUser{Username: "user1"}
		err := repo.FirstOrCreate(ctx, map[string]interface{}{
			"username": "user1",
		}, user)
		assert.NoError(t, err)
		assert.Equal(t, "user1@example.com", user.Email) // 应该找到已存在的
	})

	t.Run("记录不存在，创建新记录", func(t *testing.T) {
		user := &TestUser{
			Username: "newuser",
			Email:    "new@example.com",
			Age:      30,
		}
		err := repo.FirstOrCreate(ctx, map[string]interface{}{
			"username": "newuser",
		}, user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})
}

// TestRepo_Transaction 测试事务
func TestRepo_Transaction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[TestUser](db)
	ctx := context.Background()

	t.Run("事务成功提交", func(t *testing.T) {
		err := repo.Transaction(ctx, func(ctx context.Context, db *gorm.DB, txRepo IRepo[TestUser]) error {
			user1 := &TestUser{Username: "tx_user1", Email: "tx1@example.com"}
			if err := txRepo.Create(ctx, user1); err != nil {
				return err
			}

			user2 := &TestUser{Username: "tx_user2", Email: "tx2@example.com"}
			if err := txRepo.Create(ctx, user2); err != nil {
				return err
			}

			return nil
		})
		assert.NoError(t, err)

		// 验证两条都创建成功
		count, _ := repo.Count(ctx, nil)
		assert.Equal(t, int64(2), count)
	})

	t.Run("事务回滚", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewRepo[TestUser](db)

		err := repo.Transaction(ctx, func(ctx context.Context, db *gorm.DB, txRepo IRepo[TestUser]) error {
			user := &TestUser{Username: "rollback_user", Email: "rollback@example.com"}
			if err := txRepo.Create(ctx, user); err != nil {
				return err
			}

			// 返回错误触发回滚
			return assert.AnError
		})
		assert.Error(t, err)

		// 验证记录未创建
		count, _ := repo.Count(ctx, nil)
		assert.Equal(t, int64(0), count)
	})
}
