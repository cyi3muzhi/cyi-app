package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

//	入参为5 输出以下
//	   *
//	  ***
//	 *****
//	  ***
//	   *
//
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

func worker(ctx context.Context, url string) error {
	if url == "cyi.com" {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(7 * time.Second):
			return nil // 真没人打断才睡满 7 秒
		}
	}
	if url == "muz.com" {
		time.Sleep(3 * time.Second)
		// err := errors.New("muz.com error")
		return nil
	}
	fmt.Println("url = ", url)

	return nil
}

// 并发任务执行器
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	g, ectx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(3)
	urls := []string{"cyi.com", "muz.com", "cin.com", "cyi3muz.com"}
	fmt.Println("调用前:", runtime.NumGoroutine())
	start := time.Now()
	for _, url := range urls {
		u := url
		g.Go(func() error {
			if err := sem.Acquire(ectx, 1); err != nil {
				return err
			}
			defer sem.Release(1)
			return worker(ectx, u)
		})
	}
	fmt.Println("调用后:", runtime.NumGoroutine())

	if err := g.Wait(); err != nil {
		fmt.Println("error", err)
		fmt.Println("耗时:", time.Since(start))
		return
	}
	fmt.Println("全部完成！")
}
