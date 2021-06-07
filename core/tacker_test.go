package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTacker(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	testSites, err := filepath.Glob(filepath.Join(filepath.Dir(filename), "tests", "*"))
	if err != nil {
		t.Fatal("unable to read test sites")
	}

	for _, site := range testSites {
		tacker, err := NewTacker(site)
		assert.NoError(t, err)

		assert.NoError(t, tacker.Tack())
		AssertDirEquals(t, filepath.Join(site, "output.expected"), filepath.Join(site, "output"))
	}
}

func AssertDirEquals(t *testing.T, expected string, result string) {
	expList, err := FindFiles(expected)
	if err != nil {
		t.Errorf("unable to read dir %s: %s", expected, err)
	}

	for idx, fn := range expList {
		expList[idx] = strings.TrimPrefix(fn, expected)
	}

	resList, err := FindFiles(result)
	if err != nil {
		t.Errorf("unable to read dir %s: %s", result, err)
	}

	for idx, fn := range resList {
		resList[idx] = strings.TrimPrefix(fn, result)
	}

	if !StringSliceEqual(expList, resList) {
		t.Errorf("File lists differ: %s <--> %s", expected, result)
	}

	assert.Equal(t, expList, resList)

	for _, fn := range expList {
		expContent, err := os.ReadFile(expected + fn)
		if err != nil {
			t.Errorf("unable to read file %s: %s", fn, err)
		}
		resContent, err := os.ReadFile(result + fn)
		if err != nil {
			t.Errorf("unable to read file %s: %s", fn, err)
		}
		assert.Equal(t, expContent, resContent, "file content does not match for %s", fn)
	}
}

func StringSliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for idx, val := range a {
		if b[idx] != val {
			return false
		}
	}

	return true
}
