package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("PRYX_TEST_MODE", "true")
	os.Setenv("PRYX_DATA_DIR", "/tmp/pryx-test-data")

	os.MkdirAll("/tmp/pryx-test-data", 0755)

	code := m.Run()

	os.RemoveAll("/tmp/pryx-test-data")

	os.Exit(code)
}
