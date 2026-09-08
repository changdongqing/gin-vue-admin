package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/service"
	"github.com/gin-gonic/gin"
)

// 采集平台 API（03 文档 §六）。路由挂在 GVA 私有组（JWTAuth + CasbinHandler），
// path 形如 /collect/**（与 report 模块同惯例，无 /api 前缀，casbin 按 path+method 判权）。

type CollectApi struct{}

var (
	channelService  = service.ChannelService{}
	deviceService   = service.DeviceService{}
	variableService = service.VariableService{}
	chainService    = service.ParseChainService{}
	realtimeService = service.RealtimeService{}
)

func operatorID(c *gin.Context) uint {
	if claims, ok := c.Get("claims"); ok {
		if custom, ok := claims.(*request.CustomClaims); ok {
			return custom.BaseClaims.ID
		}
	}
	return 0
}

// ---------- 通道 ----------

// CreateChannel 创建通道。
func (CollectApi) CreateChannel(c *gin.Context) {
	var ch model.CollectChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := channelService.CreateChannel(&ch); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(ch, c)
}

// UpdateChannel 更新通道。
func (CollectApi) UpdateChannel(c *gin.Context) {
	var ch model.CollectChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := channelService.UpdateChannel(&ch); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteChannel 删除通道（级联 + 下线链）。
func (CollectApi) DeleteChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := channelService.DeleteChannel(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	service.RemoveChannelChain(uint(id))
	response.OkWithMessage("删除成功", c)
}

// FindChannel 通道详情。
func (CollectApi) FindChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ch, err := channelService.FindChannel(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(ch, c)
}

// GetChannelList 通道列表。
func (CollectApi) GetChannelList(c *gin.Context) {
	list, err := channelService.GetChannelList()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetTree 通道-设备-测点树。
func (CollectApi) GetTree(c *gin.Context) {
	tree, err := channelService.GetTree()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(tree, c)
}

// ---------- 设备 ----------

// CreateDevice 创建设备。
func (CollectApi) CreateDevice(c *gin.Context) {
	var d model.CollectDevice
	if err := c.ShouldBindJSON(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := deviceService.CreateDevice(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(d, c)
}

// UpdateDevice 更新设备。
func (CollectApi) UpdateDevice(c *gin.Context) {
	var d model.CollectDevice
	if err := c.ShouldBindJSON(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := deviceService.UpdateDevice(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteDevice 删除设备。
func (CollectApi) DeleteDevice(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := deviceService.DeleteDevice(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetDeviceList 设备列表（?channelId=）。
func (CollectApi) GetDeviceList(c *gin.Context) {
	channelID, _ := strconv.Atoi(c.Query("channelId"))
	list, err := deviceService.GetDeviceListByChannel(uint(channelID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ---------- 测点 ----------

// CreateVariable 创建测点。
func (CollectApi) CreateVariable(c *gin.Context) {
	var v model.CollectVariable
	if err := c.ShouldBindJSON(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := variableService.CreateVariable(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(v, c)
}

// UpdateVariable 更新测点。
func (CollectApi) UpdateVariable(c *gin.Context) {
	var v model.CollectVariable
	if err := c.ShouldBindJSON(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := variableService.UpdateVariable(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteVariable 删除测点。
func (CollectApi) DeleteVariable(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := variableService.DeleteVariable(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetVariableList 测点列表（?deviceId=）。
func (CollectApi) GetVariableList(c *gin.Context) {
	deviceID, _ := strconv.Atoi(c.Query("deviceId"))
	list, err := variableService.GetVariableListByDevice(uint(deviceID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ---------- 部署 ----------

// DeployChannel 编译部署通道（幂等）。
func (CollectApi) DeployChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	skipped, err := service.DeployChannel(uint(id), operatorID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if skipped {
		response.OkWithMessage("配置无变更，跳过部署", c)
		return
	}
	response.OkWithMessage("部署成功", c)
}

// UndeployChannel 下线通道。
func (CollectApi) UndeployChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.UndeployChannel(uint(id), operatorID(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已下线", c)
}

// RebuildAll 全量重编译。
func (CollectApi) RebuildAll(c *gin.Context) {
	total, ok, skipped, failures := service.RebuildAll(operatorID(c))
	response.OkWithData(gin.H{
		"total": total, "ok": ok, "skipped": skipped, "failures": failures,
	}, c)
}

// GetDeployments 部署历史（?channelId=）。
func (CollectApi) GetDeployments(c *gin.Context) {
	channelID, _ := strconv.Atoi(c.Query("channelId"))
	db := service.DB()
	var list []model.CollectDeployment
	q := db.Order("id desc").Limit(50)
	if channelID != 0 {
		q = q.Where("channel_id = ?", channelID)
	}
	if err := q.Find(&list).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ---------- 子流程库 ----------

// CreateParseChain 创建子流程。
func (CollectApi) CreateParseChain(c *gin.Context) {
	var pc model.CollectParseChain
	if err := c.ShouldBindJSON(&pc); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := chainService.CreateParseChain(&pc); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(pc, c)
}

// UpdateParseChain 更新子流程。
func (CollectApi) UpdateParseChain(c *gin.Context) {
	var pc model.CollectParseChain
	if err := c.ShouldBindJSON(&pc); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := chainService.UpdateParseChain(&pc); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteParseChain 删除子流程。
func (CollectApi) DeleteParseChain(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := chainService.DeleteParseChain(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetParseChainList 子流程列表。
func (CollectApi) GetParseChainList(c *gin.Context) {
	list, err := chainService.GetParseChainList()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// PublishParseChain 发布子流程。
func (CollectApi) PublishParseChain(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := chainService.PublishParseChain(uint(id), operatorID(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("发布成功", c)
}

// TestParseChain 子流程回放测试：注入 test_payload 同步执行并返回输出。
func (CollectApi) TestParseChain(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	out, err := chainService.TestParseChain(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}

// ---------- 实时数据与状态 ----------

// GetRealtime 实时值查询。
func (CollectApi) GetRealtime(c *gin.Context) {
	var q service.RealtimeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := realtimeService.GetRealtime(q)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetChannelStatus 通道运行状态。
func (CollectApi) GetChannelStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	st, err := realtimeService.GetChannelStatus(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(st, c)
}
