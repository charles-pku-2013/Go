// +build OMIT

package main

import "fmt"

func main() {
	// 数组格式: [len]type{values}
	primes := [6]int{2, 3, 5, 7, 11, 13}

	// 切片格式: var name []type = array[range]
	var s []int = primes[1:4]
	fmt.Println(s)
}
