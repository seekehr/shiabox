package vector

import (
	"context"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
)

func Connect() (qdrant.PointsClient, error) {
	conn, err := grpc.NewClient("localhost:6333")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := qdrant.NewPointsClient(conn)
	ctx := context.Background()
	collectionName := "shiabox"

	_, err = qdrant.NewCollectionsClient(conn).Create(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     1536,
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}
	return client, nil
}
