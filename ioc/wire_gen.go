// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitApp() (*App, error) {
	viper := InitViper()
	cmdable := InitRedis(viper)
	provider := InitSession(viper, cmdable)
	v := InitGinMiddlewares()
	mongo := InitMongoDB()
	module, err := model.InitModule(mongo)
	if err != nil {
		return nil, err
	}
	handler := module.Hdl
	attributeModule, err := attribute.InitModule(mongo)
	if err != nil {
		return nil, err
	}
	webHandler := attributeModule.Hdl
	resourceModule, err := resource.InitModule(mongo, attributeModule)
	if err != nil {
		return nil, err
	}
	handler2 := resourceModule.Hdl
	relationModule, err := relation.InitModule(mongo, attributeModule, resourceModule)
	if err != nil {
		return nil, err
	}
	relationModelHandler := relationModule.RMHdl
	relationResourceHandler := relationModule.RRHdl
	relationTypeHandler := relationModule.RTHdl
	config := InitLdapConfig(viper)
	userModule, err := user.InitModule(mongo, config)
	if err != nil {
		return nil, err
	}
	handler3 := userModule.Hdl
	engine := InitWebServer(provider, v, handler, webHandler, handler2, relationModelHandler, relationResourceHandler, relationTypeHandler, handler3)
	app := &App{
		Web: engine,
	}
	return app, nil
}

// wire.go:

var BaseSet = wire.NewSet(InitViper, InitMongoDB, InitRedis)
