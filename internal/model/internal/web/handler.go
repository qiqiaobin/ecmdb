package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	svc         service.Service
	mgSvc       service.MGService
	resourceSvc resource.Service
	RMSvc       relation.RMSvc
	AttrSvc     attribute.Service
}

func NewHandler(svc service.Service, groupSvc service.MGService, RMSvc relation.RMSvc, attrSvc attribute.Service,
	resourceSvc resource.Service) *Handler {
	return &Handler{
		svc:         svc,
		mgSvc:       groupSvc,
		AttrSvc:     attrSvc,
		RMSvc:       RMSvc,
		resourceSvc: resourceSvc,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/model")
	// 模型分组
	g.POST("/group/create", ginx.WrapBody[CreateModelGroupReq](h.CreateGroup))
	g.POST("/group/list", ginx.WrapBody[Page](h.ListModelGroups))
	g.POST("/group/delete", ginx.WrapBody[DeleteModelGroup](h.DeleteModelGroup))

	// 模型操作
	g.POST("/create", ginx.WrapBody[CreateModelReq](h.CreateModel))
	g.POST("/detail", ginx.WrapBody[DetailModelReq](h.DetailModel))
	g.POST("/list", ginx.WrapBody[Page](h.ListModels))

	// 获取模型并分组
	g.POST("/list/pipeline", ginx.WrapBody[Page](h.ListModelsByGroupId))

	g.POST("/delete", ginx.WrapBody[DeleteModelByUidReq](h.DeleteModelByUid))
	// 模型关联关系
	g.POST("/relation/diagram", ginx.WrapBody[Page](h.FindRelationModelDiagram))
	g.POST("/relation/graph", ginx.WrapBody[Page](h.FindRelationModelGraph))
}

func (h *Handler) CreateGroup(ctx *gin.Context, req CreateModelGroupReq) (ginx.Result, error) {
	id, err := h.mgSvc.CreateModelGroup(ctx.Request.Context(), domain.ModelGroup{
		Name: req.Name,
	})

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型分组成功",
	}, nil
}

func (h *Handler) CreateModel(ctx *gin.Context, req CreateModelReq) (ginx.Result, error) {
	id, err := h.svc.CreateModel(ctx, domain.Model{
		Name:    req.Name,
		GroupId: req.GroupId,
		UID:     req.UID,
		Icon:    req.Icon,
	})

	if err != nil {
		return systemErrorResult, err
	}

	_, err = h.AttrSvc.CreateDefaultAttribute(ctx, req.UID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "添加模型成功",
	}, nil
}

func (h *Handler) DetailModel(ctx *gin.Context, req DetailModelReq) (ginx.Result, error) {
	m, err := h.svc.FindModelById(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: m,
		Msg:  "模型查找成功",
	}, nil
}

func (h *Handler) ListModelGroups(ctx *gin.Context, req Page) (ginx.Result, error) {
	mgs, total, err := h.mgSvc.ListModelGroups(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelGroupsListResp{
			Total: total,
			Mgs: slice.Map(mgs, func(idx int, m domain.ModelGroup) ModelGroup {
				return h.toModelGroupVo(m)
			}),
		},
	}, nil
}

func (h *Handler) DeleteModelByUid(ctx *gin.Context, req DeleteModelByUidReq) (ginx.Result, error) {
	// TODO errorGroup改造
	_, mTotal, err := h.RMSvc.ListModelUidRelation(ctx, 0, 1, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	if mTotal != 0 {
		return modelRelationIsNotFountResult, fmt.Errorf("模型关联不为空")
	}

	_, rTotal, err := h.resourceSvc.ListResource(ctx, nil, req.ModelUid, 0, 1)
	if err != nil {
		return systemErrorResult, err
	}

	if rTotal != 0 {
		return modelRelationIsNotFountResult, fmt.Errorf("模型关联资产数据不为空")
	}

	// TODO 删除模型，同步删除模型属性 +  模型属性分组
	count, err := h.svc.DeleteModelByUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
		Msg:  "删除模型成功",
	}, nil
}

func (h *Handler) ListModelsByGroupId(ctx *gin.Context, req Page) (ginx.Result, error) {
	var (
		eg            errgroup.Group
		mgs           []domain.ModelGroup
		models        []domain.Model
		total         int64
		resourceCount map[string]int
	)

	eg.Go(func() error {
		var err error
		mgs, total, err = h.mgSvc.ListModelGroups(ctx, req.Offset, req.Limit)
		if err != nil {
			return err
		}
		mgids := make([]int64, total)
		mgids = slice.Map(mgs, func(idx int, src domain.ModelGroup) int64 {
			return src.ID
		})

		models, err = h.svc.ListModelByGroupIds(ctx, mgids)
		if err != nil {
			return err
		}

		return err
	})

	eg.Go(func() error {
		var err error
		resourceCount, err = h.resourceSvc.PipelineByModelUid(ctx)
		if err != nil {
			return err
		}

		return err
	})

	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}

	var mds map[int64][]Model
	mds = slice.ToMapV(models, func(m domain.Model) (int64, []Model) {
		return m.GroupId, slice.FilterMap(models, func(idx int, src domain.Model) (Model, bool) {
			if m.GroupId == src.GroupId {
				count := 0
				if val, ok := resourceCount[src.UID]; ok {
					count = val
				}

				return Model{
					Id:    src.ID,
					Name:  src.Name,
					Total: count,
					UID:   src.UID,
					Icon:  src.Icon,
				}, true

			}
			return Model{}, false
		})
	})

	return ginx.Result{
		Data: RetrieveModelListByGroupId{
			Mgs: slice.Map(mgs, func(idx int, src domain.ModelGroup) ModelListByGroupId {
				mb := ModelListByGroupId{}
				val, ok := mds[src.ID]
				if ok {
					mb.Models = val
				}

				mb.GroupName = src.Name
				mb.GroupId = src.ID
				return mb
			}),
		},
	}, nil
}

func (h *Handler) ListModels(ctx *gin.Context, req Page) (ginx.Result, error) {
	models, total, err := h.svc.ListModels(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveModelsListResp{
			Total: total,
			Models: slice.Map(models, func(idx int, m domain.Model) Model {
				return toModelVo(m)
			}),
		},
	}, nil
}

func (h *Handler) DeleteModelGroup(ctx *gin.Context, req DeleteModelGroup) (ginx.Result, error) {
	count, err := h.mgSvc.DeleteModelGroup(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) FindRelationModelDiagram(ctx *gin.Context, req Page) (ginx.Result, error) {
	// TODO 为了后续加入 label 概念进行过滤先查询所有的模型
	// 查询所有模型
	models, _, err := h.svc.ListModels(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	// 取出所有的 uids
	modelUids := slice.Map(models, func(idx int, src domain.Model) string {
		return src.UID
	})

	// 查询包含的数据
	ds, err := h.RMSvc.FindModelDiagramBySrcUids(ctx, modelUids)
	if err != nil {
		return systemErrorResult, err
	}

	// 生成关联节点的map
	var mds map[string][]relation.ModelDiagram
	mds = slice.ToMapV(ds, func(m relation.ModelDiagram) (string, []relation.ModelDiagram) {
		return m.SourceModelUid, slice.FilterMap(ds, func(idx int, src relation.ModelDiagram) (relation.ModelDiagram, bool) {
			if m.SourceModelUid == src.SourceModelUid {
				return relation.ModelDiagram{
					ID:              src.ID,
					RelationTypeUid: src.RelationTypeUid,
					TargetModelUid:  src.TargetModelUid,
					SourceModelUid:  src.SourceModelUid,
				}, true
			}
			return relation.ModelDiagram{}, false
		})
	})

	// 返回 vo，前端展示
	diagrams := toModelDiagramVo(models, mds)

	return ginx.Result{
		Data: RetrieveRelationModelDiagram{
			Diagrams: diagrams,
		},
	}, nil
}

func (h *Handler) FindRelationModelGraph(ctx *gin.Context, req Page) (ginx.Result, error) {
	// TODO 为了后续加入 label 概念进行过滤先查询所有的模型
	// 查询所有模型
	models, _, err := h.svc.ListModels(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	mn := make([]ModelNode, len(models))
	mn = slice.Map(models, func(idx int, src domain.Model) ModelNode {
		data := make(map[string]string, 1)
		data["icon"] = src.Icon
		return ModelNode{
			ID:   src.UID,
			Text: src.Name,
			Data: data,
		}
	})

	// 取出所有的 uids
	modelUids := slice.Map(models, func(idx int, src domain.Model) string {
		return src.UID
	})

	// 查询包含的数据
	ds, err := h.RMSvc.FindModelDiagramBySrcUids(ctx, modelUids)
	if err != nil {
		return systemErrorResult, err
	}

	ml := make([]ModelLine, len(ds))
	ml = slice.Map(ds, func(idx int, src relation.ModelDiagram) ModelLine {
		return ModelLine{
			From: src.SourceModelUid,
			To:   src.TargetModelUid,
			Text: src.RelationTypeUid,
		}
	})

	return ginx.Result{
		Data: RetrieveRelationModelGraph{
			Nodes:  mn,
			Lines:  ml,
			RootId: "virtual",
		},
	}, nil
}

func (h *Handler) toModelGroupVo(m domain.ModelGroup) ModelGroup {
	return ModelGroup{
		Name: m.Name,
		Id:   m.ID,
	}
}
