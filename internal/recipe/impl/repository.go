package impl

import (
	"context"
	"errors"
	"flove/job/config"
	"flove/job/internal/base/database"
	"flove/job/internal/recipe"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	recipesCollection = "recipes"
)

type recipeEntity struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"`
	Name        string              `bson:"name"`
	Description string              `bson:"description"`
	Category    string              `bson:"category"`
	Tags        []string            `bson:"tags"`
	Nutrition   nutritionInfoEntity `bson:"nutrition_info"`
	Servings    int                 `bson:"servings"`
	CreatedAt   time.Time           `bson:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at"`
}

type nutritionInfoEntity struct {
	Calories      float64 `bson:"calories"`
	Protein       float64 `bson:"protein"`
	Fat           float64 `bson:"fat"`
	Carbohydrates float64 `bson:"carbohydrates"`
	Fiber         float64 `bson:"fiber"`
	Sugar         float64 `bson:"sugar"`
	Sodium        float64 `bson:"sodium"`
}

func (e *recipeEntity) toRecipeModel() *recipe.RecipeModel {
	return &recipe.RecipeModel{
		ID:          e.ID.Hex(),
		Name:        e.Name,
		Description: e.Description,
		Category:    e.Category,
		Tags:        e.Tags,
		Nutrition: recipe.NutritionInfo{
			Calories:      e.Nutrition.Calories,
			Protein:       e.Nutrition.Protein,
			Fat:           e.Nutrition.Fat,
			Carbohydrates: e.Nutrition.Carbohydrates,
			Fiber:         e.Nutrition.Fiber,
			Sugar:         e.Nutrition.Sugar,
			Sodium:        e.Nutrition.Sodium,
		},
		Servings:  e.Servings,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func toEntity(r *recipe.RecipeModel) *recipeEntity {
	return &recipeEntity{
		Name:        r.Name,
		Description: r.Description,
		Category:    r.Category,
		Tags:        r.Tags,
		Nutrition: nutritionInfoEntity{
			Calories:      r.Nutrition.Calories,
			Protein:       r.Nutrition.Protein,
			Fat:           r.Nutrition.Fat,
			Carbohydrates: r.Nutrition.Carbohydrates,
			Fiber:         r.Nutrition.Fiber,
			Sugar:         r.Nutrition.Sugar,
			Sodium:        r.Nutrition.Sodium,
		},
		Servings:  r.Servings,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

type repository struct {
	config *config.Config
	db     *mongo.Database
}

func NewRecipeRepository(config *config.Config, db *mongo.Database) recipe.RecipeRepository {
	ctx := context.Background()

	db.Collection(recipesCollection).Indexes().DropAll(ctx)
	db.Collection(recipesCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"name": "text"},
	})

	return &repository{
		config: config,
		db:     db,
	}
}

func (repo *repository) CreateRecipe(ctx context.Context, r *recipe.RecipeModel) error {
	result, err := repo.db.Collection(recipesCollection).InsertOne(ctx, toEntity(r))
	if err != nil {
		return err
	}

	r.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (repo *repository) DeleteRecipe(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	result, err := repo.db.Collection(recipesCollection).DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return database.ErrNotFound
	}

	return nil
}

func (repo *repository) GetRecipeByID(ctx context.Context, id string) (*recipe.RecipeModel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, database.ErrNotFound
	}

	filter := bson.M{"_id": objectID}
	var entity *recipeEntity

	if err := repo.db.Collection(recipesCollection).FindOne(ctx, filter).Decode(entity); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}

	return entity.toRecipeModel(), nil
}

func (repo *repository) IncrementLikes(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return database.ErrNotFound
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$inc": bson.M{"likes": 1}}

	result, err := repo.db.Collection(recipesCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return database.ErrNotFound
	}

	return nil
}

// IncrementViews implements recipe.RecipeRepository.
func (repo *repository) IncrementViews(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return database.ErrNotFound
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$inc": bson.M{"views": 1}}

	result, err := repo.db.Collection(recipesCollection).UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return database.ErrNotFound
	}

	return nil
}

func (repo *repository) SearchRecipe(ctx context.Context, query string, tags []string, page, limit int64) ([]*recipe.RecipeModel, int, error) {
	skip := (max(1, page) - 1) * limit

	opts := options.
		Find().
		SetLimit(limit).
		SetSkip(skip)

	query = strings.TrimSpace(query)
	filter := bson.M{}

	if query != "" {
		filter["$text"] = bson.M{"$search": query}
		opts = opts.
			SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}).
			SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})
	}

	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	cursor, err := repo.db.Collection(recipesCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	var results []*recipeEntity
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	recipes := make([]*recipe.RecipeModel, len(results))
	for i, r := range results {
		recipes[i] = r.toRecipeModel()
	}

	count, err := repo.db.Collection(recipesCollection).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	totalDocuments := int(count)

	return recipes, totalDocuments, nil
}

func (repo *repository) UpdateRecipe(ctx context.Context, id string, update recipe.UpdateRecipeDTO) error {
	panic("unimplemented")
}
