// internal/domain/infrahub.go
package domain

import "io"

// InfrahubClient defines methods for interacting with Infrahub.
type InfrahubClient interface {
	Login(apiURL, username, password string) (string, error)
	RunQuery(queryName string, apiURL string, artifactName string, targetBranche string, targetDate string, token string) (*[]Artifact, error)
	// BuildURL(apiURL, path string, queryParams, headers map[string]string) (string, error)
	DownloadArtifact(apiURL string, artifactID string, targetBranche string, targetDate string) (io.Reader, error)
}
