package main

import "fmt"

//  入参为5 输出以下
//     *
//    ***
//   *****
//    ***
//     *
// printDiamond 打印指定大小的菱形图案
// 参数 n: 菱形的高度，必须为奇数
func printDiamond(n int) {
	// 计算中间行的位置
	mid := n / 2

	// 打印上半部分（包括中间行）
	for i := 0; i <= mid; i++ {
		// 打印空格
		for j := 0; j < mid-i; j++ {
			fmt.Print(" ")
		}
		// 打印星号
		for j := 0; j < 2*i+1; j++ {
			fmt.Print("*")
		}
		fmt.Println()
	}

	// 打印下半部分
	for i := mid - 1; i >= 0; i-- {
		// 打印空格
		for j := 0; j < mid-i; j++ {
			fmt.Print(" ")
		}
		// 打印星号
		for j := 0; j < 2*i+1; j++ {
			fmt.Print("*")
		}
		fmt.Println()
	}
}

func main() {
	// 调用函数，传入5作为参数
	printDiamond(5)
}
