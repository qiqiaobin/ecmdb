// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package department

import (
	"github.com/Duke1616/ecmdb/internal/department/internal/repository"
	"github.com/Duke1616/ecmdb/internal/department/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/department/internal/service"
	"github.com/Duke1616/ecmdb/internal/department/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitModule(db *mongox.Mongo) (*Module, error) {
	departmentDAO := dao.NewDepartmentDAO(db)
	departmentRepository := repository.NewDepartmentRepository(departmentDAO)
	serviceService := service.NewService(departmentRepository)
	handler := web.NewHandler(serviceService)
	module := &Module{
		Hdl: handler,
		Svc: serviceService,
	}
	return module, nil
}

// wire.go:

var ProviderSet = wire.NewSet(web.NewHandler, service.NewService, repository.NewDepartmentRepository, dao.NewDepartmentDAO)
