package health

import "context"

type Checker interface {
	Name() string
	Check(ctx context.Context) Result
}

type FuncChecker struct {
	name string
	fn   func(ctx context.Context) Result
}

func NewFuncChecker(name string, fn func(ctx context.Context) Result) *FuncChecker {
	return &FuncChecker{name: name, fn: fn}
}

func (c *FuncChecker) Name() string {
	return c.name
}

func (c *FuncChecker) Check(ctx context.Context) Result {
	return c.fn(ctx)
}
