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
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitApp() (*App, error) {
	v := InitGinMiddlewares()
	client := InitMongoDB()
	handler := model.InitHandler(client)
	module, err := attribute.InitModule(client)
	if err != nil {
		return nil, err
	}
	webHandler := module.Hdl
	resourceModule, err := resource.InitModule(client, module)
	if err != nil {
		return nil, err
	}
	handler2 := resourceModule.Hdl
	relationModule, err := relation.InitModule(client, module, resourceModule)
	if err != nil {
		return nil, err
	}
	relationModelHandler := relationModule.RMHdl
	relationResourceHandler := relationModule.RRHdl
	relationTypeHandler := relationModule.RTHdl
	engine := InitWebServer(v, handler, webHandler, handler2, relationModelHandler, relationResourceHandler, relationTypeHandler)
	app := &App{
		Web: engine,
	}
	return app, nil
}

// wire.go:

var BaseSet = wire.NewSet(InitMongoDB)
