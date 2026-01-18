package db

import (
	"context"
	"time"

	"github.com/supperdoggy/spot-models"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

type Database interface {
	NewDownloadRequest(ctx context.Context, url, name string, creatorID int64) error
	GetActiveRequests(ctx context.Context) ([]models.DownloadQueueRequest, error)
	DeactivateRequest(ctx context.Context, id string) error
	NewPlaylistRequest(ctx context.Context, url string, creatorID int64, noPull bool) error
}

type db struct {
	conn *mongo.Client
	log  *zap.Logger

	// Collections
	downloadQueueRequestCollection *mongo.Collection
	playlistRequestCollection      *mongo.Collection
}

func NewDatabase(ctx context.Context, log *zap.Logger, url, dbname string) (Database, error) {
	conn, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	return &db{
		conn: conn,
		log:  log,

		downloadQueueRequestCollection: conn.Database(dbname).Collection("download-queue-requests"),
		playlistRequestCollection:      conn.Database(dbname).Collection("playlist-requests"),
	}, nil
}

func (d *db) NewDownloadRequest(ctx context.Context, url, name string, creatorID int64) error {
	id := uuid.NewV4()
	request := models.DownloadQueueRequest{
		SpotifyURL: url,
		Name:       name,
		Active:     true,
		ID:         id.String(),
		CreatedAt:  time.Now().Unix(),
		CreatorID:  creatorID,
	}

	_, err := d.downloadQueueRequestCollection.InsertOne(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (d *db) NewPlaylistRequest(ctx context.Context, url string, creatorID int64, noPull bool) error {
	id := uuid.NewV4()
	request := models.PlaylistRequest{
		SpotifyURL: url,
		Active:     true,
		ID:         id.String(),
		CreatedAt:  time.Now().Unix(),
		CreatorID:  creatorID,
		NoPull:     noPull,
	}

	_, err := d.playlistRequestCollection.InsertOne(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (d *db) GetActiveRequests(ctx context.Context) ([]models.DownloadQueueRequest, error) {
	var requests []models.DownloadQueueRequest

	cursor, err := d.downloadQueueRequestCollection.Find(ctx, bson.M{"active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var request models.DownloadQueueRequest
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	return requests, nil
}

func (d *db) DeactivateRequest(ctx context.Context, id string) error {
	_, err := d.downloadQueueRequestCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"active": false, "updated_at": time.Now().Unix()}})
	if err != nil {
		return err
	}

	return nil
}
