package db

import (
	"context"
	"fmt"
	"time"

	models "github.com/supperdoggy/spot-models"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Database interface {
	NewDownloadRequest(ctx context.Context, url, name string, creatorID int64) error
	GetActiveRequests(ctx context.Context) ([]models.DownloadQueueRequest, error)
	DeactivateRequest(ctx context.Context, id string) error
	NewPlaylistRequest(ctx context.Context, url string, creatorID int64, noPull bool) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
}

type db struct {
	conn *mongo.Client
	log  *zap.Logger

	// Collections
	downloadQueueRequestCollection *mongo.Collection
	playlistRequestCollection      *mongo.Collection
}

func NewDatabase(ctx context.Context, log *zap.Logger, url, dbname string) (Database, error) {
	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := conn.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &db{
		conn: conn,
		log:  log,

		downloadQueueRequestCollection: conn.Database(dbname).Collection("download-queue-requests"),
		playlistRequestCollection:      conn.Database(dbname).Collection("playlist-requests"),
	}, nil
}

func (d *db) Close(ctx context.Context) error {
	return d.conn.Disconnect(ctx)
}

func (d *db) Ping(ctx context.Context) error {
	return d.conn.Ping(ctx, nil)
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
		return fmt.Errorf("failed to insert download request: %w", err)
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
		return fmt.Errorf("failed to insert playlist request: %w", err)
	}

	return nil
}

func (d *db) GetActiveRequests(ctx context.Context) ([]models.DownloadQueueRequest, error) {
	var requests []models.DownloadQueueRequest

	cursor, err := d.downloadQueueRequestCollection.Find(ctx, bson.M{"active": true})
	if err != nil {
		return nil, fmt.Errorf("failed to find active requests: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &requests); err != nil {
		return nil, fmt.Errorf("failed to decode requests: %w", err)
	}

	return requests, nil
}

func (d *db) DeactivateRequest(ctx context.Context, id string) error {
	result, err := d.downloadQueueRequestCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"active": false, "updated_at": time.Now().Unix()}},
	)
	if err != nil {
		return fmt.Errorf("failed to deactivate request: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("request with id %s not found", id)
	}

	return nil
}
