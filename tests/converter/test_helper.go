package converter_test

import (
	"path/filepath"
	"runtime"
)

func getTestDataPath(filename string) string {
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..")
	return filepath.Join(projectRoot, "testdata", filename)
}

