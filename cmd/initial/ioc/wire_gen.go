// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/google/wire"
)

import (
	_ "github.com/go-sql-driver/mysql"
)

// Injectors from wire.go:

func InitApp() (*App, error) {
	mongo := InitMongoDB()
	client := InitRediSearch()
	config := InitLdapConfig()
	db := InitMySQLDB()
	syncedEnforcer := InitCasbin(db)
	module, err := policy.InitModule(syncedEnforcer)
	if err != nil {
		return nil, err
	}
	departmentModule, err := department.InitModule(mongo)
	if err != nil {
		return nil, err
	}
	userModule, err := user.InitModule(mongo, client, config, module, departmentModule)
	if err != nil {
		return nil, err
	}
	service := userModule.Svc
	roleModule, err := role.InitModule(mongo)
	if err != nil {
		return nil, err
	}
	serviceService := roleModule.Svc
	dao := version.NewDao(mongo)
	versionService := version.NewService(dao)
	app := &App{
		UserSvc: service,
		RoleSvc: serviceService,
		VerSvc:  versionService,
	}
	return app, nil
}

// wire.go:

var BaseSet = wire.NewSet(InitMongoDB, InitMySQLDB, InitRedis, InitRediSearch, InitMQ, InitEtcdClient, InitLdapConfig)
