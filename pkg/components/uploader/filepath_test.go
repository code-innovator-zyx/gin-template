package uploader

import (
	"encoding/json"
	"testing"
)

func setUp(t *testing.M) {

}

// TestFilePath_Deserialization 测试FilePath的反序列化功能
func TestFilePath_Deserialization(t *testing.T) {
	jsonStr := `{"image":"http://localhost:8080/uploads/images/2024-01-01/test.jpg"}`

	type TestStruct struct {
		Image FilePath `json:"image"`
	}

	var data TestStruct
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 验证反序列化后存储的是相对路径
	if data.Image.String() != "images/2024-01-01/test.jpg" {
		t.Errorf("Expected relative path 'images/2024-01-01/test.jpg', got '%s'", data.Image.String())
	}

	t.Logf("Deserialized relative path: %s", data.Image.String())
}
