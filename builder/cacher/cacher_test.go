package cacher_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OpenBazaar/mason/builder/cacher"
)

func mustGetCleanTempDir(dirMemo string) (string, func()) {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(r.Intn(9999)) * time.Nanosecond)
	var (
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
		p, clean              = mustGetCleanTempDir("cacher-cachepath")
		expectedStore         = "preparedStoreName"
		expectedVersion       = "preparedVersionName"
		expectedPath          = filepath.Join(p, expectedStore, "sampleBinary")
		binaryPath            = filepath.Join(buildPath, "sampleBinary")
	)
	defer clean()
	defer buildClean()
	mustCreateTestBinary(binaryPath)

	c, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := c.Get("nonExistantStoreName", "nonExistantVersion"); err == nil {
		t.Errorf("expected getting non-existant store to return error, but did not")
	} else {
		if err != cacher.ErrNoStoreFound {
			t.Errorf("expected ErrNoStoreFound but got %s", err.Error())
		}
	}

	if err := c.Cache(expectedStore, expectedVersion, binaryPath); err != nil {
		t.Fatalf("expected path cache return success, but returned error: %s", err.Error())
	}

	if _, err := c.Get(expectedStore, "nonExistantVersion"); err == nil {
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

func TestCacherPersistsAndReadsIndex(t *testing.T) {
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

	persistingCache, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	if err := persistingCache.Cache(expectedStore, expectedVersion, binaryPath); err != nil {
		t.Fatalf("expected path cache return success, but returned error: %s", err.Error())
	}

	readingCache, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	path, err := readingCache.Get(expectedStore, expectedVersion)
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

	if err := validateCacheIndexHasVersion(indexPath, expectedVersion); err != nil {
		t.Error(err)
	}
}

func validateCacheIndexHasVersion(indexPath, version string) error {
	indexBytes, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return err
	}

	var cacheIndex = make(map[string]string)
	if err := json.Unmarshal(indexBytes, &cacheIndex); err != nil {
		return err
	}

	if _, ok := cacheIndex[version]; !ok {
		return fmt.Errorf("version (%s) not found in index", version)
	}
	return nil
}

func TestCanCacheMultiple(t *testing.T) {
	var (
		newTestBinary = func() (string, func()) {
			var (
				buildPath, buildClean = mustGetCleanTempDir("cacher-buildpath")
				r                     = rand.New(rand.NewSource(time.Now().UnixNano()))
				binaryPath            = filepath.Join(buildPath, fmt.Sprintf("binary%d", r.Intn(9999)))
			)
			mustCreateTestBinary(binaryPath)
			return binaryPath, buildClean
		}
		p, clean     = mustGetCleanTempDir("cacher-storepath")
		namespace    = "binaryname"
		c, err       = cacher.OpenOrCreate(p)
		testVersions = make([]string, 0)
	)
	defer clean()

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		var (
			path, cleanup = newTestBinary()
			testVersion   = fmt.Sprintf("v%d", i)
		)
		if err := c.Cache(namespace, testVersion, path); err != nil {
			t.Fatalf("caching (%s): %s", path, err.Error())
		}
		testVersions = append(testVersions, testVersion)
		defer cleanup()
	}

	d, err := cacher.OpenOrCreate(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range testVersions {
		path, err := d.Get(namespace, version)
		if err != nil {
			t.Errorf("getting version (%s): %s", version, err.Error())
		}
		if len(path) == 0 {
			t.Errorf("path for version (%s) was empty", version)

		}
	}
}
