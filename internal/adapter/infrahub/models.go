package infrahub

import (
	"github.com/infrahub-operator/vidra/internal/domain"
)

// API Request and Response Models

type queryPayload struct {
	Variables map[string]string `json:"variables"`
}

type artifactIDQueryResult struct {
	Data struct {
		CoreArtifact struct {
			Edges []struct {
				Node struct {
					ID        string `json:"id"`
					StorageID struct {
						ID string `json:"id"`
					} `json:"storage_id"`
					Checksum struct {
						Value string `json:"value"`
					} `json:"checksum"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"CoreArtifact"`
	} `json:"data"`
}

// CreateArtifactsFromAPIResponse maps an API response to a slice of domain.Artifact
func CreateArtifactsFromAPIResponse(apiResponse artifactIDQueryResult) []domain.Artifact {
	// If no artifacts are found in the API response
	if len(apiResponse.Data.CoreArtifact.Edges) == 0 {
		return nil
	}

	var artifacts = make([]domain.Artifact, 0, len(apiResponse.Data.CoreArtifact.Edges))

	// Iterate over all edges in the API response
	for _, edge := range apiResponse.Data.CoreArtifact.Edges {
		node := edge.Node

		// Create the domain.Artifact and append it to the slice
		artifact := domain.Artifact{
			ID:        node.ID,
			StorageID: node.StorageID.ID,
			Checksum:  node.Checksum.Value,
		}

		artifacts = append(artifacts, artifact)
	}

	return artifacts
}

type LoginResponse struct {
	Token string `json:"access_token"`
}
