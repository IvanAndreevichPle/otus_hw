package main

import (
	"errors"
	"os"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name     string
		offset   int64
		limit    int64
		wantFile string
	}{
		{"all", 0, 0, "testdata/out_offset0_limit0.txt"},
		{"limit10", 0, 10, "testdata/out_offset0_limit10.txt"},
		{"limit1000", 0, 1000, "testdata/out_offset0_limit1000.txt"},
		{"limit10000", 0, 10000, "testdata/out_offset0_limit10000.txt"},
		{"offset100_limit1000", 100, 1000, "testdata/out_offset100_limit1000.txt"},
		{"offset6000_limit1000", 6000, 1000, "testdata/out_offset6000_limit1000.txt"},
		{"offset500000_limit1000", 6000, 1000, "testdata/out_offset6000_limit1000.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := "out.txt"
			err := Copy("testdata/input.txt", out, tt.offset, tt.limit)
			if err != nil {
				t.Fatalf("Copy() error = %v", err)
			}
			got, _ := os.ReadFile(out)
			want, _ := os.ReadFile(tt.wantFile)
			if string(got) != string(want) {
				t.Errorf("Files differ for %s", tt.name)
			}
			os.Remove(out)
		})
	}
}

func TestCopyOfExceedsFileSize(t *testing.T) {
	err := Copy("testdata/out_offset0_limit1000.txt", "out.txt", 99999999999999, 10)
	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Errorf("expected ErrOffsetExceedsFileSize, got %v", err)
	}
}

func TestCopyFileNotExist(t *testing.T) {
	err := Copy("notExist.txt", "out.txt", 0, 0)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got %v", err)
	}
}

func TestCopyFromDirectory(t *testing.T) {
	dir := t.TempDir()

	err := Copy(dir, "out.txt", 0, 0)
	if err != ErrUnsupportedFile {
		t.Errorf("expected ErrUnsupportedFile, got %v", err)
	}
}
