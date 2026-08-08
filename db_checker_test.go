package health

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

type testPingDriver struct{}
type testPingConn struct{ err error }

var testPingErr error

func init() {
	sql.Register("health_test_ping", &testPingDriver{})
}

func (d *testPingDriver) Open(string) (driver.Conn, error) {
	return &testPingConn{err: testPingErr}, nil
}

func (c *testPingConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *testPingConn) Close() error {
	return nil
}

func (c *testPingConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c *testPingConn) Ping(ctx context.Context) error {
	return c.err
}

func TestDBCheckerName(t *testing.T) {
	checker := NewDBChecker("db", nil)
	if checker.Name() != "db" {
		t.Fatalf("expected name db, got %s", checker.Name())
	}
}

func TestDBCheckerCheckSuccess(t *testing.T) {
	testPingErr = nil
	db, err := sql.Open("health_test_ping", "")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	checker := NewDBChecker("db", db)
	res := checker.Check(context.Background())

	if res.Status != StatusUp {
		t.Fatalf("expected up status, got %s", res.Status)
	}
	if res.Type != "database" {
		t.Fatalf("expected database type, got %s", res.Type)
	}
}

func TestDBCheckerCheckFailure(t *testing.T) {
	testPingErr = errors.New("ping failed")
	db, err := sql.Open("health_test_ping", "")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	checker := NewDBChecker("db", db)
	res := checker.Check(context.Background())

	if res.Status != StatusDown {
		t.Fatalf("expected down status, got %s", res.Status)
	}
	if res.Type != "database" {
		t.Fatalf("expected database type, got %s", res.Type)
	}
	if res.Error != "ping failed" {
		t.Fatalf("expected ping failed, got %q", res.Error)
	}
}
