package health

import (
	"context"
	"strings"
	"syscall"
	"testing"
)

func TestNewDiskCheckerDefaults(t *testing.T) {
	checker := NewDiskChecker("", "", 0)

	if checker.Name() != "disk_space" {
		t.Fatalf("expected default name disk_space, got %s", checker.Name())
	}
	if checker.path != "/" {
		t.Fatalf("expected default path /, got %s", checker.path)
	}
}

func TestDiskCheckerCheckReportsLowDiskSpace(t *testing.T) {
	checker := NewDiskChecker("disk", "/", ^uint64(0))
	res := checker.Check(context.Background())

	if res.Status != StatusDown {
		t.Fatalf("expected down status, got %s", res.Status)
	}
	if !strings.Contains(res.Error, "low disk space") {
		t.Fatalf("expected low disk space error, got %q", res.Error)
	}
	if res.Type != "storage" {
		t.Fatalf("expected storage type, got %s", res.Type)
	}
}

func TestDiskCheckerCheckHandlesInvalidPath(t *testing.T) {
	checker := NewDiskChecker("disk", "/definitely/does/not/exist", 0)
	res := checker.Check(context.Background())

	if res.Status != StatusDown {
		t.Fatalf("expected down status, got %s", res.Status)
	}
	if !strings.Contains(res.Error, "failed to get disk stats") {
		t.Fatalf("expected disk stats error, got %q", res.Error)
	}
}

func TestDiskCheckerCheckReportsHealthyStateWhenEnoughSpace(t *testing.T) {
	originalStatfs := statfs
	t.Cleanup(func() {
		statfs = originalStatfs
	})

	statfs = func(path string, stat *syscall.Statfs_t) error {
		stat.Bavail = 0
		stat.Bfree = 0
		stat.Blocks = 0
		stat.Bsize = 1024
		return nil
	}

	checker := NewDiskChecker("disk", "/tmp", 0)
	res := checker.Check(context.Background())

	if res.Status != StatusUp {
		t.Fatalf("expected up status, got %s", res.Status)
	}
	if res.Type != "storage" {
		t.Fatalf("expected storage type, got %s", res.Type)
	}
	if got := res.Details["free_percentage"]; got != "0.00%" {
		t.Fatalf("expected 0.00%% free percentage, got %v", got)
	}
}

func TestHumanizeBytes(t *testing.T) {
	cases := []struct {
		name  string
		input uint64
		want  string
	}{
		{name: "bytes", input: 42, want: "42 B"},
		{name: "kilobytes", input: 1536, want: "1.50 KB"},
		{name: "megabytes", input: 3 * 1024 * 1024, want: "3.00 MB"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanizeBytes(tt.input); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
