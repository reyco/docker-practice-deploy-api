package models

type User struct {
	username, password, pet string
}

var (
	ModelSettingsUser = &ModelSettings{"/users", "User", "users", User{}}
)
