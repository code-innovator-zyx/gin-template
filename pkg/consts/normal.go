package consts

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/14 下午5:58
* @Package:  通用枚举类型
 */

// Gender 性别
type Gender uint8

const (
	GenderUnknown Gender = iota
	GenderMale
	GenderFemale
)

// String 实现 fmt.Stringer 接口
func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "男"
	case GenderFemale:
		return "女"
	case GenderUnknown:
		fallthrough
	default:
		return "-"
	}
}
func AllGender() []Gender {
	return []Gender{GenderMale, GenderFemale}
}
