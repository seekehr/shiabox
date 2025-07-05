package vector

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"server/internal/constants"
	"strconv"
)

const (
	CollectionName = "shiabox"
)

type Db struct {
	Client qdrant.PointsClient
}

func Connect() (*Db, error) {
	conn, err := grpc.NewClient(
		"localhost:6334", // 6333 is the http server port
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	client := qdrant.NewPointsClient(conn)
	ctx := context.Background()

	collClient := qdrant.NewCollectionsClient(conn)

	_, err = collClient.Get(ctx, &qdrant.GetCollectionInfoRequest{CollectionName: CollectionName})
	if err != nil {
		// only create collection if it's not found
		if status.Code(err) == codes.NotFound {
			_, err = collClient.Create(ctx, &qdrant.CreateCollection{
				CollectionName: CollectionName,
				VectorsConfig: &qdrant.VectorsConfig{
					Config: &qdrant.VectorsConfig_Params{
						Params: &qdrant.VectorParams{
							Size:     1024,
							Distance: qdrant.Distance_Cosine,
						},
					},
				},
			})

			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &Db{
		Client: client,
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
	upsert, err := db.Client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: CollectionName,
		Points:         ahadithAsPoints,
	})

	if err != nil {
		return err
	}
	getStatus := upsert.GetResult().GetStatus()
	if getStatus != qdrant.UpdateStatus_Acknowledged && getStatus != qdrant.UpdateStatus_Completed {
		return fmt.Errorf("error adding ahadith to vector db. status: %d", getStatus)
	}

	return nil
}
