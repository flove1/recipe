package recommendation

const (
	VIEWED = iota
	LIKED
	SAVED
)

type RecipeModel struct {
	Name     string
	Category string
	Tags     []string
}
