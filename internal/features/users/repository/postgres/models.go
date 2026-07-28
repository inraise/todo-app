package postgres_repository

type UserModel struct {
	ID          int
	Version     int
	FullName    string
	PhoneNumber *string
}
