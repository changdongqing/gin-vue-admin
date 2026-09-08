package model

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// 采集平台三级模型（aiDoc/rulego/03-采集平台快速采集配置详细设计 §三）。
// 三级模型为唯一事实源，规则链 DSL 为编译派生物。

// CollectChannel 通道：一条物理/逻辑连接 + 驱动（对应 ThingsGateway.Channel）。
type CollectChannel struct {
	global.GVA_MODEL
	Name         string         `json:"name" gorm:"comment:通道名称;uniqueIndex" binding:"required"`
	AccessMode   string         `json:"accessMode" gorm:"size:16;comment:接入方式 poll=轮询采集 report=主动上报" binding:"required,oneof=poll report"`
	Driver       string         `json:"driver" gorm:"size:32;comment:驱动标识，仅poll通道使用 modbus/bacnet/s7/opcua/snmp/fins/mc/iec104/dlt645/eip；report通道置空"`
	ConnConfig   datatypes.JSON `json:"connConfig" gorm:"type:jsonb;comment:连接配置JSON(结构随驱动,字段对齐组件配置元数据)"`
	PollInterval int            `json:"pollInterval" gorm:"comment:通道默认轮询周期ms"`
	OutputConfig datatypes.JSON `json:"outputConfig" gorm:"type:jsonb;comment:大JSON输出配置{mqttPublish:{server,username,password,topic,qos},dbWrite:{enable}}"`
	Enable       *bool          `json:"enable" gorm:"default:true;comment:通道使能"`
	Remark       string         `json:"remark" gorm:"size:255;comment:备注"`
}

func (CollectChannel) TableName() string { return "collect_channels" }

// CollectDevice 设备：通道下的采集单元（对应 ThingsGateway.Device）。
type CollectDevice struct {
	global.GVA_MODEL
	ChannelID    uint           `json:"channelId" gorm:"comment:所属通道ID" binding:"required"`
	Name         string         `json:"name" gorm:"size:64;index:idx_collect_device_name;comment:设备名称(通道内唯一,service层查重)"`
	DeviceKind   string         `json:"deviceKind" gorm:"size:16;comment:设备类别 register=寄存器型 report=报文型" binding:"required,oneof=register report"`
	DeviceTypeID *uint          `json:"deviceTypeId" gorm:"comment:设备类型ID(报文型必填)"`
	UnitID       *int           `json:"unitId" gorm:"comment:寄存器型站号"`
	PollInterval int            `json:"pollInterval" gorm:"comment:设备级周期ms(0=用通道默认)"`
	Props        datatypes.JSON `json:"props" gorm:"type:jsonb;comment:驱动扩展属性"`
	Enable       *bool          `json:"enable" gorm:"default:true;comment:设备使能"`
	Remark       string         `json:"remark" gorm:"size:255;comment:备注"`
}

func (CollectDevice) TableName() string { return "collect_devices" }

// CollectVariable 测点：设备下的点位（对应 ThingsGateway.Variable）。
type CollectVariable struct {
	global.GVA_MODEL
	DeviceID     uint    `json:"deviceId" gorm:"comment:所属设备ID" binding:"required"`
	Name         string  `json:"name" gorm:"size:64;index:idx_collect_variable_name;comment:测点名(设备内唯一,service层查重)"`
	Addr         string  `json:"addr" gorm:"size:64;comment:协议地址或报文取数路径"`
	DataType     string  `json:"dataType" gorm:"size:16;comment:数据类型"`
	Scale        float64 `json:"scale" gorm:"comment:线性换算系数(0=不换算)"`
	Offset       float64 `json:"offset" gorm:"comment:线性换算偏移(0=无偏移)"`
	Endian       string  `json:"endian" gorm:"size:8;comment:字节序 ABCD/CDAB/BADC/DCBA"`
	CollectGroup string  `json:"collectGroup" gorm:"size:32;comment:采集分组(同组同批读取)"`
	ReadExpr     string  `json:"readExpr" gorm:"size:255;comment:复杂读表达式(一期仅透传子流程契约)"`
	RW           string  `json:"rw" gorm:"size:8;comment:读写 R/RW"`
	Unit         string  `json:"unit" gorm:"size:16;comment:工程单位"`
	Enable       *bool   `json:"enable" gorm:"default:true;comment:测点使能"`
	Remark       string  `json:"remark" gorm:"size:255;comment:备注"`
}

func (CollectVariable) TableName() string { return "collect_variables" }

// CollectDeviceType 设备类型（报文型）：绑定解析子流程。
type CollectDeviceType struct {
	global.GVA_MODEL
	Name         string `json:"name" gorm:"size:64;comment:设备类型名;uniqueIndex"`
	PayloadType  string `json:"payloadType" gorm:"size:8;comment:输入契约 string/json/bytes"`
	ParseChainID *uint  `json:"parseChainId" gorm:"comment:绑定子流程ID"`
	Enable       *bool  `json:"enable" gorm:"default:true"`
	Remark       string `json:"remark" gorm:"size:255;comment:备注"`
}

func (CollectDeviceType) TableName() string { return "collect_device_types" }

// CollectParseChain 子流程库：开发标准化交付的解析规则链。
type CollectParseChain struct {
	global.GVA_MODEL
	Name          string `json:"name" gorm:"size:64;comment:子流程名;uniqueIndex"`
	Version       int    `json:"version" gorm:"comment:语义版本(发布递增)"`
	Dsl           string `json:"dsl" gorm:"type:text;comment:规则链DSL"`
	InputContract string `json:"inputContract" gorm:"size:255;comment:输入契约说明"`
	TestPayload   string `json:"testPayload" gorm:"type:text;comment:测试报文样例"`
	Status        string `json:"status" gorm:"size:16;comment:draft/published"`
	Remark        string `json:"remark" gorm:"size:255;comment:备注"`
}

func (CollectParseChain) TableName() string { return "collect_parse_chains" }

// CollectDeployment 部署记录：通道编译部署历史（幂等依据 dsl_hash）。
type CollectDeployment struct {
	global.GVA_MODEL
	ChannelID       uint   `json:"channelId" gorm:"index;comment:通道ID"`
	ChainID         string `json:"chainId" gorm:"size:64;comment:rulego链ID"`
	DslHash         string `json:"dslHash" gorm:"size:64;comment:编译产物SHA-256"`
	DslSnapshot     string `json:"dslSnapshot" gorm:"type:text;comment:编译快照"`
	TemplateVersion string `json:"templateVersion" gorm:"size:16;comment:编译模板版本"`
	Status          string `json:"status" gorm:"size:16;comment:deployed/rolled_back"`
	OperatorID      uint   `json:"operatorId" gorm:"comment:操作人ID"`
}

func (CollectDeployment) TableName() string { return "collect_deployments" }

// CollectRealtime 测点最新值（采集链 UPSERT，无 GVA_MODEL——由 point_key 定位）。
type CollectRealtime struct {
	PointKey  string `json:"pointKey" gorm:"primaryKey;size:128;comment:通道:设备:测点"`
	ChannelID uint   `json:"channelId" gorm:"index;comment:通道ID"`
	DeviceID  uint   `json:"deviceId" gorm:"index;comment:设备ID"`
	Name      string `json:"name" gorm:"size:64;comment:测点名"`
	Value     string `json:"value" gorm:"comment:最新值(字符串化)"`
	Quality   string `json:"quality" gorm:"size:16;comment:good/bad/timeout"`
	ErrorMsg  string `json:"errorMsg" gorm:"size:255;comment:错误信息"`
	Ts        int64  `json:"ts" gorm:"comment:采集时间戳(unix ms)"`
}

func (CollectRealtime) TableName() string { return "collect_realtime" }
