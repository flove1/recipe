package recipe

type UpdateRecipeDTO struct {
	Name        *string
	Description *string
	Category    *string
	Tags        *[]string
	Nutrition   *NutritionInfo
	CookTime    *int
	Servings    *int
}
