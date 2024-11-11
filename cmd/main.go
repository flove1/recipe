package main

import (
	"context"
	"flove/job/config"
	"flove/job/internal/api/http"
	"flove/job/internal/auth"
	"flove/job/internal/base/database"
	"flove/job/internal/recipe"
	"flove/job/internal/recommendation"
	"flove/job/internal/user"
	"log"
	"os"
	"os/signal"
	"strings"

	authImpl "flove/job/internal/auth/impl"
	recipeImpl "flove/job/internal/recipe/impl"
	recommendationImpl "flove/job/internal/recommendation/impl"
	userImpl "flove/job/internal/user/impl"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// @title           Recipe API
// @version         0.0.1
// @description     API for recipe recommendation application

// @contact.name   Kirill Shaforostov
// @contact.email  dragon090986@gmail.com

// @host      localhost:8080

// @securitydefinitions.apikey ApiKeyAuth
// @in cookie
// @name Authorization

func subscribeToRecipes(eventBus *database.EventBus, neo4jDriver neo4j.DriverWithContext) {
	eventBus.Subscribe("recipe:created", func(message string) {
		input := strings.Split(message, ":")
		recipeID := input[0]
		name := input[1]
		category := input[2]
		tags := strings.Split(input[3], ",")

		session := neo4jDriver.NewSession(context.TODO(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		_, err := session.ExecuteWrite(context.TODO(),
			func(tx neo4j.ManagedTransaction) (any, error) {
				query := `CREATE (r:Recipe {recipeID: $id, name: $name, category: $category, tags: $tags})`
				params := map[string]any{
					"id":       recipeID,
					"name":     name,
					"category": category,
					"tags":     tags,
				}

				_, err := tx.Run(context.TODO(), query, params)
				return nil, err
			})

		if err != nil {
			log.Printf("Error creating recipe node in Neo4j: %v", err)
		} else {
			log.Printf("Successfully created recipe node in Neo4j")
		}

		session.Close(context.TODO())
	})

	eventBus.Subscribe("recipe:deleted", func(message string) {
		session := neo4jDriver.NewSession(context.TODO(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		_, err := session.ExecuteWrite(context.TODO(),
			func(tx neo4j.ManagedTransaction) (any, error) {
				query := `MATCH (r:Recipe {id: $id}) DELETE r`
				params := map[string]any{"id": message}

				_, err := tx.Run(context.TODO(), query, params)
				return nil, err
			})

		if err != nil {
			log.Printf("Error deleting recipe node in Neo4j: %v", err)
		} else {
			log.Printf("Successfully deleted recipe node in Neo4j")
		}
	})
}

func subscribeToUsers(eventBus *database.EventBus, neo4jDriver neo4j.DriverWithContext) {
	eventBus.Subscribe("user:created", func(message string) {
		input := strings.Split(message, ":")
		userID := input[0]

		session := neo4jDriver.NewSession(context.TODO(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		_, err := session.ExecuteWrite(context.TODO(),
			func(tx neo4j.ManagedTransaction) (any, error) {
				query := `CREATE (u:User {userID: $id})`
				params := map[string]any{
					"id": userID,
				}

				_, err := tx.Run(context.TODO(), query, params)
				return nil, err
			})

		if err != nil {
			log.Printf("Error creating user node in Neo4j: %v", err)
		} else {
			log.Printf("Successfully created user node in Neo4j")
		}

		session.Close(context.TODO())
	})

	eventBus.Subscribe("user:deleted", func(message string) {
		session := neo4jDriver.NewSession(context.TODO(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		_, err := session.ExecuteWrite(context.TODO(),
			func(tx neo4j.ManagedTransaction) (any, error) {
				query := `MATCH (r:User {userID: $id}) DETACH DELETE r`
				params := map[string]any{"id": message}

				_, err := tx.Run(context.TODO(), query, params)
				return nil, err
			})

		if err != nil {
			log.Printf("Error deleting user node in Neo4j: %v", err)
		} else {
			log.Printf("Successfully deleted user node in Neo4j")
		}
	})
}

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		panic(err)
	}

	mongoClient, err := database.NewMongoConnection(cfg.Mongo.URL)
	if err != nil {
		panic(err)
	}

	defer mongoClient.Disconnect(context.Background())
	mongoDB := mongoClient.Database(cfg.Mongo.Name)

	redisClient, err := database.NewRedisConnection(cfg.Redis.URL)
	if err != nil {
		panic(err)
	}

	defer redisClient.Close()

	neo4jDriver, err := database.NewNeo4jConnection(cfg.Neo4j.URL)
	if err != nil {
		panic(err)
	}

	defer neo4jDriver.Close(context.Background())

	eventBus := database.NewRedisEventBus(redisClient)
	subscribeToRecipes(eventBus, neo4jDriver)
	subscribeToUsers(eventBus, neo4jDriver)

	userRepo := userImpl.NewUserRepository(cfg, mongoDB)
	userUC := userImpl.NewUserUC(cfg, eventBus, userRepo)
	userHandler := user.NewUserHandler(userUC)

	refreshTokenRepo := authImpl.NewRefreshTokenRepository(cfg, mongoDB)
	accessTokenRepo := authImpl.NewAccessTokenRepository(cfg, redisClient)
	tokenUC := authImpl.NewTokenUC(cfg, accessTokenRepo, refreshTokenRepo, userRepo)
	authHandler := auth.NewTokenHandler(tokenUC, userUC)

	recipeRepo := recipeImpl.NewRecipeRepository(cfg, mongoDB)
	recipeUC := recipeImpl.NewRecipeUC(cfg, eventBus, recipeRepo)
	recipeHandler := recipe.NewRecipeHandler(cfg, recipeUC)

	recommendationRepo := recommendationImpl.NewRecommendationRepository(cfg, neo4jDriver)
	recommendationUC := recommendationImpl.NewRecommendationUC(cfg, eventBus, recommendationRepo)
	recommendationHandler := recommendation.NewRecommendationHandler(cfg, recommendationUC)

	server := http.NewServer(cfg, http.Handlers{
		UserHandler:           userHandler,
		TokenHandler:          authHandler,
		RecipeHandler:         recipeHandler,
		RecommendationHandler: recommendationHandler,
	})
	server.Start()
	log.Println("server started")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case s := <-interrupt:
		log.Printf("signal received: %s", s.String())
	case err = <-server.Notify():
		log.Printf("server notify: %s", err.Error())
	}

	err = server.Shutdown()
	if err != nil {
		log.Printf("server shutdown err: %s", err)
	}

	log.Println("Server exiting")
}
