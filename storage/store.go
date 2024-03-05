package storage

import "errors"

var (
	ErrNotFound = errors.New("object not found")
)

type Store interface {
	Get(string) (PackageMetadata, error)
	// Put(Package, []byte) error
}
