package handle

import "net/http"

const (
	AbbreviatedPackageMetadataContentType = "application/vnd.npm.install-v1+json"
)

func PackageMetadata(w http.ResponseWriter, r *http.Request) (int, error) {
	return http.StatusOK, nil
}
