package vector

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"server/internal/constants"
	"strconv"
)

const (
	collectionName = "shiabox"
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

func GenerateUUID(hadith constants.HadithEmbedding) string {
	fullBytes := md5.Sum([]byte(hadith.Book + strconv.Itoa(hadith.Hadith)))
	return hex.EncodeToString(fullBytes[:])
}

func (db *Db) Add(vector []constants.HadithEmbedding) error {
	ahadithAsPoints := make([]*qdrant.PointStruct, len(vector))
	for i, hadith := range vector {
		qdrantPoint := &qdrant.PointStruct{
			Id:      qdrant.NewID(GenerateUUID(hadith)),
			Vectors: qdrant.NewVectors(hadith.Embedding...), // confusing syntax ngl ;/. this expands a slice into a variadic or whatever args
			Payload: map[string]*qdrant.Value{
				"Book":   {Kind: &qdrant.Value_StringValue{StringValue: hadith.Book}},
				"Page":   {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(hadith.Page)}},
				"Hadith": {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(hadith.Hadith)}},
			},
		}
		ahadithAsPoints[i] = qdrantPoint
	}
	upsert, err := db.client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         ahadithAsPoints,
	})

	if err != nil {
		return err
	}
	status := upsert.GetResult().GetStatus()
	if status != qdrant.UpdateStatus_Acknowledged && status != qdrant.UpdateStatus_Completed {
		return fmt.Errorf("error adding ahadith to vector db. status: %d", status)
	}

	return nil
}
