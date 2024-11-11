package user

type UpdateUserDTO struct {
	Role  *Role   `bson:"role,omitempty"`
	Phone *string `bson:"phone,omitempty"`
}
