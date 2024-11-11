package impl

import (
	"context"
	"flove/job/config"
	"flove/job/internal/recommendation"
	"flove/job/pkg/fp"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type repository struct {
	cfg    *config.Config
	driver neo4j.DriverWithContext
}

func (r *repository) NewInteraction(ctx context.Context, userID string, recipeID string, interaction int) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	var (
		interactionRelation string
		weight              int
	)

	switch interaction {
	case recommendation.VIEWED:
		interactionRelation = "VIEWED"
		weight = 1
	case recommendation.LIKED:
		interactionRelation = "LIKED"
		weight = 5
	case recommendation.SAVED:
		interactionRelation = "SAVED"
		weight = 10
	default:
		panic("Invalid interaction type")
	}

	query := fmt.Sprintf(`
		MATCH (u:User {userID: $userID}), (r:Recipe {recipeID: $recipeID})
		WHERE NOT (u)-[:%s]->(r)
		LIMIT 1
		CREATE (u)-[rel:%s]->(r)
		SET rel.weight = $weight
	`, interactionRelation, interactionRelation)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{
			"userID":   userID,
			"recipeID": recipeID,
			"weight":   weight,
		})
	})

	return err
}

func (r *repository) RecalculatePreferences(ctx context.Context, userID string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {userID: $userID})-[rel:LIKED|SAVED|VIEWED]->(r:Recipe)
		UNWIND r.tags AS tag
		WITH SUM(rel.weight) as totalSum

		MATCH (u:User {userID: $userID})-[rel:LIKED|SAVED|VIEWED]->(r:Recipe)
		UNWIND r.tags AS tag
		WITH u, tag, SUM(rel.weight) AS tagScore, totalSum as totalSum
		WITH u, tag, toFloat(tagScore) / totalSum as coefficient
		with u, collect(tag) as tags, collect(coefficient) as coefficients
		set u.preference_tags = tags, u.preference_coefficients = coefficients
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{"userID": userID})
	})

	return err
}

func NewRecommendationRepository(cfg *config.Config, driver neo4j.DriverWithContext) recommendation.RecommendationRepository {
	return &repository{
		cfg:    cfg,
		driver: driver,
	}
}

func (r *repository) GetRecommendationCollaborative(ctx context.Context, userID string) ([]recommendation.RecipeModel, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {userID: $userID})-[:LIKED|SAVED|VIEWED]->(:Recipe)<-[:LIKED|SAVED|VIEWED]-(similar:User)-[interaction:LIKED|SAVED|VIEWED]->(r:Recipe)
		WHERE NOT (u)-[:LIKED|SAVED|VIEWED]->(r)
		RETURN r.name AS name, r.category as category, r.tags as tags, SUM(interaction.weight) AS weightedScore
		ORDER BY weightedScore DESC
		LIMIT 5
	`

	results, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, query, map[string]interface{}{"userID": userID})
		if err != nil {
			return nil, err
		}

		var recipes []recommendation.RecipeModel
		for records.Next(ctx) {
			record := records.Record().AsMap()
			tags := fp.Map(record["tags"].([]interface{}), func(tag any) string { return tag.(string) })

			recipes = append(recipes, recommendation.RecipeModel{
				Name:     record["name"].(string),
				Category: record["category"].(string),
				Tags:     tags,
			})
		}

		return recipes, nil
	})

	if results == nil {
		return []recommendation.RecipeModel{}, nil
	}

	return results.([]recommendation.RecipeModel), err
}

func (r *repository) GetRecommendationPreferences(ctx context.Context, userID string) ([]recommendation.RecipeModel, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {userID: $userID})
		WITH u, size(u.preference_coefficients) AS len
		UNWIND range(0, len - 1) AS i
		WITH u, u.preference_coefficients[i] AS coefficient, u.preference_tags[i] AS tag

		MATCH (r:Recipe)
		WHERE NOT (u)-[:LIKED|SAVED|VIEWED]->(r)
		UNWIND r.tags AS recipeTag
		WITH r, recipeTag, coefficient
		WHERE recipeTag = tag
		WITH r, SUM(coefficient) AS recipeScore

		RETURN r.name AS name, r.category as category, r.tags as tags
		ORDER BY recipeScore DESC
		LIMIT 5
	`

	results, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, query, map[string]interface{}{"userID": userID})
		if err != nil {
			return nil, err
		}

		var recipes []recommendation.RecipeModel
		for records.Next(ctx) {
			record := records.Record().AsMap()
			tags := fp.Map(record["tags"].([]interface{}), func(tag any) string { return tag.(string) })

			recipes = append(recipes, recommendation.RecipeModel{
				Name:     record["name"].(string),
				Category: record["category"].(string),
				Tags:     tags,
			})
		}

		return recipes, nil
	})

	if results == nil {
		return []recommendation.RecipeModel{}, nil
	}

	return results.([]recommendation.RecipeModel), err
}
