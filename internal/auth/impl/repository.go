package impl

import (
	"context"
	"errors"
	"flove/job/config"
	"flove/job/internal/auth"
	"flove/job/internal/base/database"
	"flove/job/internal/user"
	"log"
	"strconv"
	"time"

	"codnect.io/chrono"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	tokensCollection = "tokens"
)

type refreshTokenRepository struct {
	config *config.Config
	db     *mongo.Database
}

type refreshTokenEntity struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserUUID string             `bson:"user_uuid"`
	Token    string             `bson:"token"`
	Expiry   time.Time          `bson:"expires_at"`
}

func (r *refreshTokenEntity) toRefreshTokenModel() *auth.RefreshTokenModel {
	return &auth.RefreshTokenModel{
		ID:       r.ID.Hex(),
		UserUUID: r.UserUUID,
		Token:    r.Token,
		Expiry:   r.Expiry,
	}
}

func toRefreshTokenEntity(t *auth.RefreshTokenModel) *refreshTokenEntity {
	return &refreshTokenEntity{
		UserUUID: t.UserUUID,
		Token:    t.Token,
		Expiry:   t.Expiry,
	}
}

func NewRefreshTokenRepository(config *config.Config, db *mongo.Database) auth.RefreshTokenRepository {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	_, err := taskScheduler.ScheduleWithCron(func(ctx context.Context) {
		filter := bson.M{"expires_at": bson.M{"$lt": time.Now()}}

		db.Collection(tokensCollection).DeleteMany(ctx, filter)
		log.Println("expired refresh tokens are deleted")
	}, "0 0 0 * * 0")

	if err != nil {
		log.Printf("scheduling task error: %s", err.Error())
	}

	return &refreshTokenRepository{
		db:     db,
		config: config,
	}
}

func (r *refreshTokenRepository) NewRefreshToken(ctx context.Context, token *auth.RefreshTokenModel) error {
	result, err := r.db.Collection(tokensCollection).InsertOne(ctx, toRefreshTokenEntity(token))
	if err != nil {
		return err
	}

	token.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, plaintext string) (*auth.RefreshTokenModel, error) {
	filter := bson.M{"token": plaintext}
	result := &refreshTokenEntity{}

	if err := r.db.Collection(tokensCollection).FindOne(ctx, filter).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}

	return result.toRefreshTokenModel(), nil
}

func (r *refreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lt": time.Now()}}

	_, err := r.db.Collection(tokensCollection).DeleteOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return database.ErrNotFound
		}
		return err
	}

	return nil
}

func (r *refreshTokenRepository) DeleteToken(ctx context.Context, token string) error {
	filter := map[string]string{"token": token}

	_, err := r.db.Collection(tokensCollection).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

type accessTokenRepository struct {
	config *config.Config
	db     *redis.Client
}

func NewAccessTokenRepository(config *config.Config, db *redis.Client) auth.AccessTokenRepository {
	return &accessTokenRepository{
		db:     db,
		config: config,
	}
}

func (r *accessTokenRepository) NewAccessToken(ctx context.Context, token auth.AccessTokenModel) error {
	err := r.db.HSet(ctx, token.Token, "userUUID", token.UserUUID, "role", int(token.Role)).Err()
	if err != nil {
		return err
	}

	r.db.Expire(ctx, token.Token, time.Hour)

	return nil
}

func (r *accessTokenRepository) VerifyToken(ctx context.Context, token string) (string, user.Role, error) {
	userUUID, err := r.db.HGet(ctx, token, "userUUID").Result()
	if err != nil {
		return "", 0, database.ErrNotFound
	}

	result, err := r.db.HGet(ctx, token, "role").Result()
	if err != nil {
		return "", 0, database.ErrNotFound
	}

	role, err := strconv.Atoi(result)
	if err != nil {
		return "", 0, err
	}

	return userUUID, user.Role(role), nil
}
