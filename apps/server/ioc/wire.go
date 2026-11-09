//go:build wireinject
// +build wireinject

package ioc

import (
	"github.com/google/wire"

	"todo/server/config"
)

func InitializeApp(cfg *config.Config) (*App, func(), error) {
	panic(wire.Build(ProviderSet))
}
