package main

import (
	"maj-core-srvc/src/cmd"
	_ "maj-core-srvc/src/docs"
)

// @title                       Majoo Core Service API
// @version                     1.0
// @description                 RESTful API Microservice Boilerplate (Fiber + GORM).
// @BasePath                    /
// @schemes                     http https
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Masukkan JWT token dengan format: Bearer <token>
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        X-API-Key
// @description                 Internal Service-to-Service API Key
func main() {
	cmd.Execute()
}
