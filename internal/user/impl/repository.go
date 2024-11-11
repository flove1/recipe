package impl

import (
	"context"
	"errors"
	"flove/job/config"
	"flove/job/internal/base/database"
	"flove/job/internal/user"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	usersCollection = "users"
)

type userEntity struct {
	UUID         primitive.ObjectID `bson:"_id,omitempty"`
	Username     string             `bson:"username"`
	Email        string             `bson:"email"`
	Phone        string             `bson:"phone"`
	PasswordHash []byte             `bson:"password"`
	Role         user.Role          `bson:"role"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

func (e *userEntity) toUserModel() *user.UserModel {
	return &user.UserModel{
		ID:           e.UUID.Hex(),
		Username:     e.Username,
		Email:        e.Email,
		Phone:        e.Phone,
		PasswordHash: e.PasswordHash,
		Role:         e.Role,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func toEntity(u *user.UserModel) *userEntity {
	return &userEntity{
		Username:     u.Username,
		Email:        u.Email,
		Phone:        u.Phone,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

type repository struct {
	config *config.Config
	db     *mongo.Database
}

func NewUserRepository(config *config.Config, db *mongo.Database) user.UserRepository {
	return &repository{
		db:     db,
		config: config,
	}
}

func (repo *repository) CreateUser(ctx context.Context, u *user.UserModel) error {
	result, err := repo.db.Collection(usersCollection).InsertOne(ctx, toEntity(u))
	if err != nil {
		return err
	}

	u.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (repo *repository) GetUserByCredentials(ctx context.Context, email string) (*user.UserModel, error) {
	filter := map[string]string{"email": email}
	result := &userEntity{}

	err := repo.db.Collection(usersCollection).FindOne(ctx, filter).Decode(result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}

	return result.toUserModel(), nil
}

func (repo *repository) GetUserByID(ctx context.Context, id string) (*user.UserModel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, database.ErrNotFound
	}

	filter := bson.M{"_id": objectID}
	result := &userEntity{}

	if err := repo.db.Collection(usersCollection).FindOne(ctx, filter).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return result.toUserModel(), nil
}

func (repo *repository) UpdateUser(ctx context.Context, userID string, updates any) error {
	filter := bson.M{"_id": userID}
	update := map[string]any{"$set": updates}

	_, err := repo.db.Collection(usersCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) ChangeUserPassword(ctx context.Context, userID string, password string) error {
	passwordHash, err := user.HashPassword(password)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"password": passwordHash}}

	_, err = repo.db.Collection(usersCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) ChangeUserRole(ctx context.Context, userID string, role user.Role) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"role": role}}

	_, err := repo.db.Collection(usersCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) DeleteUser(ctx context.Context, userID string) error {
	filter := bson.M{"_id": userID}
	_, err := repo.db.Collection(usersCollection).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}
