package service

import (
	"gin-admin/pkg/orm"
	"testing"
	"time"

	"gorm.io/gorm"
)

// TestItem is a simple model used for repository tests.
type TestItem struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Age       int
	CreatedAt time.Time
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := orm.Init(orm.Config{
		DSN:          "root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		MaxOpenConns: 16,
		MaxIdleConns: 10,
		MaxLifetime:  3600,
		LogLevel:     2,
	})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&TestItem{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedItems(t *testing.T, db *gorm.DB, items []TestItem) {
	t.Helper()
	for i := range items {
		if err := db.Create(&items[i]).Error; err != nil {
			t.Fatalf("seed item failed: %v", err)
		}
	}
}

func TestCreateAndFindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}

	item := TestItem{Name: "Alice", Age: 30}
	if err := repo.Create(&item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.FindByID(item.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.Name != "Alice" || got.Age != 30 {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestFindByIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	seedItems(t, db, []TestItem{{Name: "A", Age: 10}, {Name: "B", Age: 20}, {Name: "C", Age: 30}})

	var ids []uint
	var all []TestItem
	if err := db.Find(&all).Error; err != nil {
		t.Fatalf("query all failed: %v", err)
	}
	for _, it := range all {
		ids = append(ids, it.ID)
	}

	got, err := repo.FindByIDs(ids[:2])
	if err != nil {
		t.Fatalf("FindByIDs failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
}

func TestExistsAndFindOne(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	seedItems(t, db, []TestItem{{Name: "Alice", Age: 18}, {Name: "Bob", Age: 22}})

	nameScope := func(name string) func(*gorm.DB) *gorm.DB {
		return func(tx *gorm.DB) *gorm.DB { return tx.Where("name = ?", name) }
	}

	exist, err := repo.Exists(nameScope("Alice"))
	if err != nil || !exist {
		t.Fatalf("Exists expected true, got exist=%v err=%v", exist, err)
	}

	one, err := repo.FindOne(nameScope("Bob"))
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if one.Name != "Bob" || one.Age != 22 {
		t.Fatalf("unexpected record: %+v", one)
	}
}

func TestFindAllWithLimitAndScope(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	seedItems(t, db, []TestItem{{Name: "A", Age: 10}, {Name: "B", Age: 20}, {Name: "C", Age: 30}})

	ageGE := func(n int) func(*gorm.DB) *gorm.DB {
		return func(tx *gorm.DB) *gorm.DB { return tx.Where("age >= ?", n) }
	}

	got, err := repo.FindAll(2, ageGE(20))
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
}

func TestUpdateByID(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	item := TestItem{Name: "Old", Age: 40}
	if err := repo.Create(&item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.UpdateByID(item.ID, map[string]any{"name": "New"}); err != nil {
		t.Fatalf("UpdateByID failed: %v", err)
	}

	got, err := repo.FindByID(item.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.Name != "New" {
		t.Fatalf("expected name 'New', got '%s'", got.Name)
	}
}

func TestUpdateSelected(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	item := TestItem{Name: "Sel", Age: 25}
	if err := repo.Create(&item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update only Age, including zero value handling by GORM Select.
	data := &TestItem{Age: 0}
	if err := repo.UpdateSelected(item.ID, []string{"Age"}, data); err != nil {
		t.Fatalf("UpdateSelected failed: %v", err)
	}

	got, err := repo.FindByID(item.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.Age != 0 || got.Name != "Sel" {
		t.Fatalf("unexpected record after UpdateSelected: %+v", got)
	}
}

func TestUpdateOmit(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	item := TestItem{Name: "Omit", Age: 33}
	if err := repo.Create(&item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data := &TestItem{Name: "OmitNew", Age: 99}
	if err := repo.UpdateOmit(item.ID, []string{"Age"}, data); err != nil {
		t.Fatalf("UpdateOmit failed: %v", err)
	}

	got, err := repo.FindByID(item.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.Name != "OmitNew" || got.Age != 33 {
		t.Fatalf("unexpected record after UpdateOmit: %+v", got)
	}
}

func TestDeleteByID(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	item := TestItem{Name: "Del", Age: 50}
	if err := repo.Create(&item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.DeleteByID(item.ID); err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	_, err := repo.FindByID(item.ID)
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	}
}

func TestListPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := &BaseRepo[TestItem]{Tx: db}
	seedItems(t, db, []TestItem{
		{Name: "N1", Age: 10},
		{Name: "N2", Age: 20},
		{Name: "N3", Age: 30},
		{Name: "N4", Age: 40},
		{Name: "N5", Age: 50},
	})

	res, err := repo.List(PageQuery{Page: 2, PageSize: 2, OrderBy: "id asc"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if res.Total != 5 {
		t.Fatalf("expected total 5, got %d", res.Total)
	}
	if res.Page != 2 || res.PageSize != 2 {
		t.Fatalf("unexpected page info: %+v", res)
	}
	if len(res.List) != 2 {
		t.Fatalf("expected 2 items on page, got %d", len(res.List))
	}

	// Verify IDs are 3 and 4 when ordered by id asc.
	// Reload to get IDs for clarity.
	var all []TestItem
	if err := db.Order("id asc").Find(&all).Error; err != nil {
		t.Fatalf("query all ordered failed: %v", err)
	}
	expectedID3 := all[2].ID
	expectedID4 := all[3].ID
	if res.List[0].ID != expectedID3 || res.List[1].ID != expectedID4 {
		t.Fatalf("unexpected page records: %+v", res.List)
	}
}
