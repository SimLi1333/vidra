package infrahub

import (
	"gitlab.ost.ch/ins-stud/sa-ba/ba-fs25-infrahub/infrahub-operator/internal/domain"
)

// API Request and Response Models

type QueryPayload struct {
	Variables map[string]string `json:"variables"`
}

type ArtifactIDQueryResult struct {
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
func CreateArtifactsFromAPIResponse(apiResponse ArtifactIDQueryResult) ([]domain.Artifact, error) {
	// If no artifacts are found in the API response
	if len(apiResponse.Data.CoreArtifact.Edges) == 0 {
		return nil, nil
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

	return artifacts, nil
}

type LoginResponse struct {
	Token string `json:"access_token"`
}
