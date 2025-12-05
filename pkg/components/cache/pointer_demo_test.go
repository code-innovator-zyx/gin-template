package cache

import (
	"context"
	"fmt"
	"testing"
)

type DemoUser struct {
	Name string
	Age  int
}

func TestPointerSetGet(t *testing.T) {
	cache := NewShardedMemoryCache(4)
	ctx := context.Background()

	t.Run("指针类型存储和获取", func(t *testing.T) {
		// Set: 存储指针
		user := &DemoUser{Name: "Alice", Age: 25}
		err := cache.Set(ctx, "user:ptr", user, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get: 获取指针
		var result DemoUser
		err = cache.Get(ctx, "user:ptr", &result)
		fmt.Printf("Get error: %v\n", err)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		fmt.Printf("Original: %+v\n", user)
		fmt.Printf("Result: %+v\n", result)

	})

	t.Run("值类型存储和获取", func(t *testing.T) {
		// Set: 存储值
		user := DemoUser{Name: "Bob", Age: 30}
		err := cache.Set(ctx, "user:val", user, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get: 获取值
		var result DemoUser
		err = cache.Get(ctx, "user:val", &result)
		fmt.Printf("Get error: %v\n", err)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		fmt.Printf("Original: %+v\n", user)
		fmt.Printf("Result: %+v\n", result)
	})

	t.Run("存储值获取指针", func(t *testing.T) {
		// Set: 存储值
		user := DemoUser{Name: "Charlie", Age: 35}
		err := cache.Set(ctx, "user:val2ptr", user, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get: 获取指针
		var result *DemoUser
		err = cache.Get(ctx, "user:val2ptr", &result)
		fmt.Printf("Get error: %v\n", err)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		fmt.Printf("Original: %+v\n", user)
		fmt.Printf("Result: %+v\n", result)
		fmt.Printf("Result is pointer: %v\n", result != nil)
	})
}
