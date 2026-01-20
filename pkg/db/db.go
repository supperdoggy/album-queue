package db

import (
	"context"
	"fmt"
	"time"

	models "github.com/supperdoggy/spot-models"
	"github.com/supperdoggy/spot-models/spotify"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Database interface {
	NewDownloadRequest(ctx context.Context, url, name string, creatorID int64, expectedTrackCount int, trackMetadata []spotify.TrackMetadata) error
	GetActiveRequests(ctx context.Context) ([]models.DownloadQueueRequest, error)
	DeactivateRequest(ctx context.Context, id string) error
	NewPlaylistRequest(ctx context.Context, url string, creatorID int64, noPull bool) error
	FindMusicFiles(ctx context.Context, artists, titles []string) ([]models.MusicFile, error)
	UpdateDownloadRequest(ctx context.Context, request models.DownloadQueueRequest) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
}

type db struct {
	conn *mongo.Client
	log  *zap.Logger

	// Collections
	downloadQueueRequestCollection *mongo.Collection
	playlistRequestCollection      *mongo.Collection
	musicFilesCollection           *mongo.Collection
	dbname                         string
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
		dbname: dbname,

		downloadQueueRequestCollection: conn.Database(dbname).Collection("download-queue-requests"),
		playlistRequestCollection:      conn.Database(dbname).Collection("playlist-requests"),
		musicFilesCollection:           conn.Database(dbname).Collection("music-files"),
	}, nil
}

func (d *db) Close(ctx context.Context) error {
	return d.conn.Disconnect(ctx)
}

func (d *db) Ping(ctx context.Context) error {
	return d.conn.Ping(ctx, nil)
}

func (d *db) NewDownloadRequest(ctx context.Context, url, name string, creatorID int64, expectedTrackCount int, trackMetadata []spotify.TrackMetadata) error {
	id := uuid.NewV4()
	request := models.DownloadQueueRequest{
		SpotifyURL:          url,
		Name:                name,
		Active:              true,
		ID:                  id.String(),
		CreatedAt:           time.Now().Unix(),
		UpdatedAt:           time.Now().Unix(),
		CreatorID:           creatorID,
		ExpectedTrackCount:  expectedTrackCount,
		FoundTrackCount:     0,
		TrackMetadata:       trackMetadata,
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

func (d *db) FindMusicFiles(ctx context.Context, artists, titles []string) ([]models.MusicFile, error) {
	if len(artists) != len(titles) {
		return nil, fmt.Errorf("artists and titles must have the same length")
	}

	orPairs := make([]bson.M, 0, len(artists))
	for i := range artists {
		orPairs = append(orPairs, bson.M{
			"artist": artists[i],
			"title":  titles[i],
		})
	}

	cur, err := d.musicFilesCollection.Find(ctx, bson.M{
		"$or": orPairs,
	}, options.Find().SetProjection(bson.M{"meta_data": 0}))
	if err != nil {
		return nil, fmt.Errorf("failed to find music files: %w", err)
	}
	defer cur.Close(ctx)

	files := make([]models.MusicFile, 0)
	for cur.Next(ctx) {
		var file models.MusicFile
		if err := cur.Decode(&file); err != nil {
			return nil, fmt.Errorf("failed to decode music file: %w", err)
		}
		files = append(files, file)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return files, nil
}

func (d *db) UpdateDownloadRequest(ctx context.Context, request models.DownloadQueueRequest) error {
	result, err := d.downloadQueueRequestCollection.UpdateOne(
		ctx,
		bson.M{"_id": request.ID},
		bson.M{"$set": bson.M{
			"expected_track_count": request.ExpectedTrackCount,
			"found_track_count":    request.FoundTrackCount,
			"track_metadata":       request.TrackMetadata,
			"name":                 request.Name,
			"active":              request.Active,
			"updated_at":          request.UpdatedAt,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update download request: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("request with id %s not found", request.ID)
	}

	return nil
}
