package rulego

// 按需引入 RuleGo 组件生态（aiDoc/rulego/03-采集平台快速采集配置详细设计 §2.3）。
//
// 采集平台的通道采集节点 x/iotRead、x/iotWrite 位于 components-iot 的
// external/deviceio 包；该包的 all.go 以同包静态导入注册全部协议驱动
//（modbus/bacnet/s7/opcua/snmp/fins/mc/iec104/dlt645/eip），同包文件无法按协议
// 拆分，因此单包引入即全量注册（03 文档风险 #6 结论修正：无法逐协议裁剪；
// 组件均为纯 Go、无 CGO，体积增量可接受）。
// MQTT 客户端（mqttClient）为 RuleGo 核心库内置节点，无需额外引入。
//
// 上游升级新增协议时无需改动本文件；升级 SOP 见 aiDoc/rulego/02。
import (
	_ "github.com/rulego/rulego-components-iot/external/deviceio"
)
