package vector

import (
	"context"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
)

type Db struct {
	client qdrant.PointsClient
}

func Connect() (*Db, error) {
	conn, err := grpc.NewClient("localhost:6333")
	if err != nil {
		return nil, err
	}
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
	return &Db{
		client: client,
	}, nil
}

func (db *Db) Add(vector []float32) (int64, error) {
	return 0, nil
}
