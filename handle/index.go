package handle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"npmeta/storage"
	"path/filepath"
	"sync"
)

var (
	ErrTarballCorrupt = errors.New("corrupt tarball")
	// mutex to control access when mutating the package mutex map
	mu = &sync.Mutex{}
	// pkgmu map of mutexes when mutating package metadata
	pkgmu = map[string]*sync.Mutex{}
)

func pkgmutex(pkg string) *sync.Mutex {
	return pkgmu[pkg]
}

func Index(w http.ResponseWriter, r *http.Request) (int, error) {
	// unmarshal body
	key, err := KeyFrom(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	// download tarball
	tgz, err := DownloadTarball(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return http.StatusNotFound, err
		}
		return http.StatusInternalServerError, err
	}
	// parse tarball
	// get package.json
	pkgj, err := PackageJsonFromTar(tgz)
	if err != nil {
		if errors.Is(err, ErrTarballCorrupt) {
			return http.StatusUnprocessableEntity, err
		}
		return http.StatusInternalServerError, err
	}
	// lock package mutex
	//    lock main mutex
	//    get or create package mutex
	// load existing metadata
	//   create metadata
	// add version metadata
	// commit to db
	// unlock package mutex
	return http.StatusOK, nil
}

func ParsePackageJson(data []byte, key string) (string, string, map[string]interface{}, error) {
	raw := map[string]interface{}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", "", raw, err
	}
	raw["dist"] = map[string]string{"tarball": key}
	version, _ := raw["version"].(string)
	name, _ := raw["name"].(string)
	return name, version, raw, nil
}

func PackageJsonFromTar(tgz []byte) ([]byte, error) {
	if len(tgz) == 0 {
		return []byte{}, fmt.Errorf("%w, can not extract package/package.json, from %s, empty data", ErrTarballCorrupt)
	}
	buffer := bytes.NewBuffer(tgz)
	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return []byte{}, fmt.Errorf("%w, error in gzip reader opening %s: %w", ErrTarballCorrupt, err)
	}

	tr := tar.NewReader(gzipReader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return []byte{}, fmt.Errorf("%w, could not find package.json", ErrTarballCorrupt)
		}
		if err != nil {
			return []byte{}, err
		}
		buf := bytes.NewBuffer([]byte{})
		matched, err := filepath.Match("*/package.json", hdr.Name)
		if err != nil {
			return []byte{}, fmt.Errorf("%w, error matching tar entry with */package.json: %w", ErrTarballCorrupt, err)
		}
		if matched {
			if _, err := io.Copy(buf, tr); err != nil {
				return []byte{}, err
			}
			return buf.Bytes(), nil
		}
	}
}

func DownloadTarball(key string) ([]byte, error) {
	resp, err := http.Get(key)
	if err != nil {
		return []byte{}, err
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return []byte{}, storage.ErrNotFound
	case http.StatusOK:
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	default:
		return []byte{}, fmt.Errorf("error calling %s responded with: %v %s", key, resp.StatusCode, resp.Status)
	}
}

func KeyFrom(r *http.Request) (string, error) {
	type IndexMsg struct {
		Key string `json:"key"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	msg := IndexMsg{}
	err = json.Unmarshal(data, &msg)
	return msg.Key, err
}
