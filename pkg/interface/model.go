package _interface

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/12/1 下午12:57
* @Package:
 */

type IModel interface {
	TableName() string
	GetID() uint
}
