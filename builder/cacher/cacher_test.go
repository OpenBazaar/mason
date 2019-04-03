package cacher_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OpenBazaar/samulator/builder/cacher"
)

func mustGetCleanTempDir(dirMemo string) (string, func()) {
	var (
		r      = rand.New(rand.NewSource(time.Now().UnixNano()))
		target = filepath.Join(os.TempDir(), fmt.Sprintf("%s_test_%d", dirMemo, r.Intn(9999)))
	)
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		if rErr := os.RemoveAll(target); rErr != nil {
			panic(fmt.Sprintf("path %s cannot be cleaned: %s", target, rErr.Error()))
		}
	}

	if err := os.MkdirAll(target, os.ModePerm); err != nil {
		panic(err)
	}
	return target, func() { os.RemoveAll(target) }
}

func mustCreateTestBinary(path string) {
	if err := ioutil.WriteFile(path, []byte("filebinary"), 0500); err != nil {
		panic(err)
	}
}

func TestCacherCacheAndGet(t *testing.T) {
	var (
		buildPath, buildClean = mustGetCleanTempDir("cacher-buildpath")
		p, clean              = mustGetCleanTempDir("cacher-cacheandget")
		expectedStore         = "store"
		expectedVersion       = "version"
		expectedPath          = filepath.Join(p, expectedStore, "binary")
		binaryPath            = filepath.Join(buildPath, "binary")
	)
	defer clean()
	defer buildClean()
	mustCreateTestBinary(binaryPath)

	c, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := c.Get("nonexistantstore", "nonexistantversion"); err == nil {
		t.Errorf("expected getting non-existant store to return error, but did not")
	} else {
		if err != cacher.ErrNoStoreFound {
			t.Errorf("expected ErrNoStoreFound but got %s", err.Error())
		}
	}

	if err := c.Cache(expectedStore, expectedVersion, binaryPath); err != nil {
		t.Fatalf("expected path cache return success, but returned error: %s", err.Error())
	}

	if _, err := c.Get(expectedStore, "nonexistantversion"); err == nil {
		t.Errorf("expected getting non-existant version to return error, but did not")
	} else {
		if err != cacher.ErrNoCacheFound {
			t.Errorf("expected ErrNoCacheFound but got %s", err.Error())
		}
	}

	path, err := c.Get(expectedStore, expectedVersion)
	if err != nil {
		t.Fatalf("expeted get to return success, but returned error: %s", err.Error())
	}

	if expectedPath != path {
		t.Errorf("expected path to be (%s), but was (%s)", expectedPath, path)
	}
}

func TestCacherPersistsReadsIndex(t *testing.T) {
	var (
		buildPath, buildClean = mustGetCleanTempDir("cacher-buildpath")
		p, clean              = mustGetCleanTempDir("cacher-persists")
		expectedStore         = "store"
		expectedVersion       = "version"
		expectedPath          = filepath.Join(p, expectedStore, "binary")
		binaryPath            = filepath.Join(buildPath, "binary")
	)
	defer clean()
	defer buildClean()
	mustCreateTestBinary(binaryPath)

	c, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Cache(expectedStore, expectedVersion, binaryPath); err != nil {
		t.Fatalf("expected path cache return success, but returned error: %s", err.Error())
	}

	d, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	path, err := d.Get(expectedStore, expectedVersion)
	if err != nil {
		t.Fatalf("expeted get to return success, but returned error: %s", err.Error())
	}

	if expectedPath != path {
		t.Errorf("expected path to be (%s), but was (%s)", expectedPath, path)
	}
}

func TestCacherCachesAndIndexes(t *testing.T) {
	var (
		buildPath, buildClean = mustGetCleanTempDir("cacher-buildpath")
		binaryPath            = filepath.Join(buildPath, "binary")
		p, clean              = mustGetCleanTempDir("cacher-storepath")
		expectedStore         = "store"
		expectedVersion       = "version"
		expectedPath          = filepath.Join(p, expectedStore, "binary")
		indexPath             = filepath.Join(p, expectedStore, ".cache_index")
	)
	defer clean()
	defer buildClean()
	mustCreateTestBinary(binaryPath)

	c, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Cache(expectedStore, expectedVersion, binaryPath); err != nil {
		t.Fatalf("expected path cache return success, but returned error: %s", err.Error())
	}

	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("expected (%s) to exist, but did not: %s", indexPath, err.Error())
	}

	path, err := c.Get(expectedStore, expectedVersion)
	if err != nil {
		t.Fatalf("expeted get to return success, but returned error: %s", err.Error())
	}

	if expectedPath != path {
		t.Errorf("expected path to be (%s), but was (%s)", expectedPath, path)
	}
}
