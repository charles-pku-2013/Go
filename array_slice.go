package main

import "fmt"

func main() {
    var a [2]string
    a[0] = "Hello"
    a[1] = "World"
    fmt.Println(a[0], a[1])
    fmt.Println(a)

    primes := [6]int{2, 3, 5, 7, 11, 13}
    fmt.Println(primes)
}

// slice
func main() {
    primes := [6]int{2, 3, 5, 7, 11, 13}

    var s []int = primes[1:4]
    fmt.Println(s)
}

// 更改切片的元素会修改其底层数组中对应的元素。
// 与它共享底层数组的切片都会观测到这些修改。
func main() {
    names := [4]string{
        "John",
        "Paul",
        "George",
        "Ringo",
    }
    fmt.Println(names)

    a := names[0:2]
    b := names[1:3]
    fmt.Println(a, b)

    b[0] = "XXX"
    fmt.Println(a, b)
    fmt.Println(names)
}

func main() {
    q := []int{2, 3, 5, 7, 11, 13}
    fmt.Println(q)

    r := []bool{true, false, true, true, false, true}
    fmt.Println(r)

    s := []struct {
        i int
        b bool
    }{
        {2, true},
        {3, false},
        {5, true},
        {7, true},
        {11, false},
        {13, true},
    }
    fmt.Println(s)
}

func main() {
    s := []int{2, 3, 5, 7, 11, 13}

    s = s[1:4]
    fmt.Println(s)

    s = s[:2]
    fmt.Println(s)

    s = s[1:]
    fmt.Println(s)
}

func main() {
    s := []int{2, 3, 5, 7, 11, 13}
    printSlice(s)

    // 截取切片使其长度为 0
    s = s[:0]
    printSlice(s)

    // 拓展其长度
    s = s[:4]
    printSlice(s)

    // 舍弃前两个值
    s = s[2:]
    printSlice(s)
}

func printSlice(s []int) {
    fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func main() {
    // make 函数会分配一个元素为零值的数组并返回一个引用了它的切片：
    a := make([]int, 5)
    printSlice("a", a)

    // 要指定它的容量，需向 make 传入第三个参数：
    b := make([]int, 0, 5)
    printSlice("b", b)

    c := b[:2]
    printSlice("c", c)

    d := c[2:5]
    printSlice("d", d)
}

func printSlice(s string, x []int) {
    fmt.Printf("%s len=%d cap=%d %v\n",
        s, len(x), cap(x), x)
}


// 切片的切片
import (
    "fmt"
    "strings"
)

func main() {
    // 创建一个井字板（经典游戏）
    board := [][]string{
        []string{"_", "_", "_"},
        []string{"_", "_", "_"},
        []string{"_", "_", "_"},
    }

    // 两个玩家轮流打上 X 和 O
    board[0][0] = "X"
    board[2][2] = "O"
    board[1][2] = "X"
    board[1][0] = "O"
    board[0][2] = "X"

    for i := 0; i < len(board); i++ {
        fmt.Printf("%s\n", strings.Join(board[i], " "))
    }
}

// 向切片追加元素
func main() {
    var s []int
    printSlice(s)

    // 添加一个空切片
    s = append(s, 0)
    printSlice(s)

    // 这个切片会按需增长
    s = append(s, 1)
    printSlice(s)

    // 可以一次性添加多个元素
    s = append(s, 2, 3, 4)
    printSlice(s)
}

// range
var pow = []int{1, 2, 4, 8, 16, 32, 64, 128}

func main() {
    for i, v := range pow {
        fmt.Printf("2**%d = %d\n", i, v)
    }
    for _, value := range pow {
        fmt.Printf("%d\n", value)
    }
}
