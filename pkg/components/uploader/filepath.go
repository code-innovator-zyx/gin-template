package uploader

import (
	"database/sql/driver"
	"encoding/json"
)

type FilePaths []FilePath

// Scan 实现 sql.Scanner 接口
func (j *FilePaths) Scan(value interface{}) error {
	if value == nil {
		*j = []FilePath{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j FilePaths) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal([]string{})
	}

	// 直接存储相对路径，避免触发 FilePath.MarshalJSON（会变成 URL）
	res := make([]string, len(j))
	for i, v := range j {
		res[i] = v.String() // 保留原始相对路径
	}
	return json.Marshal(res)
}

// FilePath 文件路径类型
// 内部存储相对路径，序列化时自动转换为完整URL
type FilePath string

// String 获取相对路径
func (fp FilePath) String() string {
	return string(fp)
}

// url 使用全局上传器实例，支持不同类型的上传器（local/OSS/S3等）
func (fp FilePath) url() string {
	return uploader.GetURL(string(fp))
}

// MarshalJSON 实现JSON序列化接口
// 序列化时自动转换为完整URL
func (fp FilePath) MarshalJSON() ([]byte, error) {
	return json.Marshal(fp.url())
}

// UnmarshalJSON 实现JSON反序列化接口
// 支持从URL或相对路径反序列化
// 使用上传器的ParseUrl方法智能解析
func (fp *FilePath) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*fp = ""
		return nil
	}
	// 使用上传器的ParseUrl方法解析（支持URL或path）
	*fp = FilePath(uploader.ParseUrl(s))
	return nil
}
