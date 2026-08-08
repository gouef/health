package health

import (
	"context"
	"database/sql"
	"time"
)

type DBChecker struct {
	name string
	db   *sql.DB
}

func NewDBChecker(name string, db *sql.DB) *DBChecker {
	return &DBChecker{name: name, db: db}
}

func (c *DBChecker) Name() string {
	return c.name
}

func (c *DBChecker) Check(ctx context.Context) Result {
	start := time.Now()
	err := c.db.PingContext(ctx)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return Result{
			Status:       StatusDown,
			Type:         "database",
			ResponseTime: duration,
			Error:        err.Error(),
		}
	}

	return Result{
		Status:       StatusUp,
		Type:         "database",
		ResponseTime: duration,
	}
}
