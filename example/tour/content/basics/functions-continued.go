// +build OMIT

package main

import "fmt"

// 函数定义格式 func 函数名(参数列表) (返回类型)
func add(x, y int) int {
	return x + y
}

func main() {
	fmt.Println(add(42, 13))
}
