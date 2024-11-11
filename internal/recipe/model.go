package recipe

import (
	"time"
)

type RecipeModel struct {
	ID          string
	Name        string
	Description string
	Category    string
	Tags        []string
	Nutrition   NutritionInfo
	Servings    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NutritionInfo struct {
	Calories      float64
	Protein       float64
	Fat           float64
	Carbohydrates float64
	Fiber         float64
	Sugar         float64
	Sodium        float64
}
