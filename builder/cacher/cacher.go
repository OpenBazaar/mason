package cacher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	logging "github.com/op/go-logging"
)

const defaultCacheStoreFilename = ".cache_index"

var log = logging.MustGetLogger("cacher")

// Cacher is used by the builder package to copy the binary of expensive builds.
type Cacher interface {
	// Cache a single binary artifact using a namespace and version
	Cache(namespace, version, path string) error
	// Get the path for an already cached binary if one exists. An error will
	// be returned for any case that causes the Cacher to not provide a valid path
	Get(namespace, version string) (string, error)
}

type cacherImpl struct {
	sync.RWMutex

	sourcePath string
	stores     map[string]cacherStore
}

type cacherStore map[string]string

func copyFile(src, dst string) error {
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !stat.Mode().IsRegular() {
		return fmt.Errorf("not regular file (%s)", src)
	}

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("cached destination exists (%s)", dst)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if err := destFile.Chmod(0755); err != nil {
		return err
	}

	buf := make([]byte, 2000)
	for {
		n, err := sourceFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destFile.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

func loadCacherStoreIndex(path string) (cacherStore, error) {
	var (
		store           cacherStore
		cacheIndex, err = ioutil.ReadFile(filepath.Join(path, defaultCacheStoreFilename))
	)
	if err != nil {
		return nil, fmt.Errorf("reading cache index (%s): %s", path, err.Error())
	}
	err = json.Unmarshal(cacheIndex, &store)
	if err != nil {
		return nil, fmt.Errorf("parsing cache index (%s): %s", path, err.Error())
	}

	return store, nil
}

func writeCacherStoreIndex(store cacherStore, path string) error {
	var sBytes, err = json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling index: %s", err.Error())
	}
	if err := ioutil.WriteFile(filepath.Join(path, defaultCacheStoreFilename), sBytes, 0644); err != nil {
		return fmt.Errorf("writing index: %s", err.Error())
	}
	return nil
}

var (
	ErrNoCacheFound = errors.New("cached version not found")
	ErrNoStoreFound = errors.New("cache store not found")
)

// OpenOrCreate expects a directory cache with populated indicies for each
// store or for no directory to exist. If a directory doesn't exist, the
// function will attempt to create one. Any non-indexed directories within
// will cause an error to be returned. A valid Cacher will be returned
// if nil error is returned.
func OpenOrCreate(path string) (*cacherImpl, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("unable to create cache (%s): %s", path, err.Error())
	}

	var dirs, err = ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open cache: %s", err.Error())
	}

	var c = &cacherImpl{
		sourcePath: path,
		stores:     make(map[string]cacherStore),
	}

	for _, dir := range dirs {
		store, err := loadCacherStoreIndex(filepath.Join(path, dir.Name()))
		if err != nil {
			return nil, fmt.Errorf("loading cache store: %s", err.Error())
		}
		c.stores[dir.Name()] = store
	}

	return c, nil
}

// Get accepts a store namespace and version string which it will use to
// locate cached versions. These strings must match exactly to return
// a cached binary
func (c *cacherImpl) Get(store, version string) (string, error) {
	c.RLock()
	defer c.RUnlock()

	s, sOK := c.stores[store]
	if !sOK {
		return "", ErrNoStoreFound
	}
	path, vOK := s[version]
	if !vOK {
		return "", ErrNoCacheFound
	}
	return path, nil
}

// Cache accepts a store namespace and version, along with the full path
// to the binary file which is to be cached. If the store namespace does
// not exist, it will be created. Any non-nil should expect the cache
// is not persisting the binary and will be returned to the prior safe state
func (c *cacherImpl) Cache(store, version, path string) error {
	var (
		storePath     = filepath.Join(c.sourcePath, store)
		cacheFilePath = filepath.Join(storePath, filepath.Base(path))
	)

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("cache target unreadable (%s): %s", path, err.Error())
	}

	if err := os.MkdirAll(storePath, 0755); err != nil {
		return fmt.Errorf("creating store path (%s): %s", storePath, err.Error())
	}

	if err := copyFile(path, cacheFilePath); err != nil {
		return fmt.Errorf("caching binary (%s -> %s): %s", path, cacheFilePath, err.Error())
	}

	c.Lock()
	defer c.Unlock()

	if _, ok := c.stores[store]; !ok {
		c.stores[store] = make(map[string]string)
	}
	c.stores[store][version] = cacheFilePath

	if err := writeCacherStoreIndex(c.stores[store], storePath); err != nil {
		log.Warningf("failed updating cache index: %s", err.Error())
		delete(c.stores[store], version)
		return fmt.Errorf("writing store cache: %s", err.Error())
	} else {
		log.Infof("updated cache index with version (%s) at (%s)", version, cacheFilePath)
	}

	return nil
}
