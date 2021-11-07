package testdata

import (
	"context"
	"fmt"
)

// func pred() bool {
// 	return true
// }

// func pp(x int) int {
// 	if x > 2 && pred() {
// 		return 5
// 	}

// 	var b = pred()
// 	if b {
// 		return 6
// 	}
// 	return 0
// }

type contextKey string

var k contextKey = "1"

// func ctxTest0() {
// 	ctx := c.WithValue(nil, k, 42)
// 	_ = ctx
// 	fmt.Println("didn't panic")
// }

// func ctxTest1() {
// 	var ctx c.Context
// 	ctx = c.WithValue(ctx, k, 42)
// 	_ = ctx
// 	fmt.Println("didn't panic")
// }

// func ctxTest2() {
// 	var ctx context.Context = nil
// 	ctx = c.WithValue(ctx, k, 42)
// 	_ = ctx
// 	fmt.Println("didn't panic")
// }

// func ctxTest3() {
// 	var ctx context.Context = nil
// 	ctx = nil
// 	ctx = c.WithValue(ctx, k, 42)
// 	_ = ctx
// 	fmt.Println("didn't panic")
// }

// func ctxInitLaterBeforeUsed() {
// 	var ctx context.Context = nil
// 	ctx = context.Background() // assign value not 'nil'
// 	ctx = context.WithValue(ctx, k, 42)
// 	_ = ctx
// 	fmt.Println("didn't panic")
// }

// func ctxInitWithBackground() {

// 	var ctx = c.Background()
// 	ctx = c.WithValue(ctx, k, 42)
// 	_ = ctx
// 	fmt.Println("didn't panic")
// }

func ctxMultiInitWithNil() {
	var ctx, ctx2 context.Context = nil, context.Background()
	ctx = context.WithValue(ctx, k, 42)
	_ = ctx
	_ = ctx2
	fmt.Println("didn't panic")
}
