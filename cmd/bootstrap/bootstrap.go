package bootstrap

import (
	"github.com/dropboks/notification-service/cmd/di"
	"github.com/dropboks/notification-service/config/env"
	"go.uber.org/dig"
)

func Run() *dig.Container {
	env.Load()
	container := di.BuildContainer()
	return container
}
