package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/router/example"
	"github.com/flipped-aurora/gin-vue-admin/server/router/ontology"
	"github.com/flipped-aurora/gin-vue-admin/server/router/report"
	"github.com/flipped-aurora/gin-vue-admin/server/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System   system.RouterGroup
	Example  example.RouterGroup
	Ontology ontology.RouterGroup
	Report   report.RouterGroup
}
