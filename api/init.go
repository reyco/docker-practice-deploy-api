package api

import (
	"../models"
)

func registerAll() {

	NewAuthService()
	NewApiService(models.ModelSettingsUser)

}
