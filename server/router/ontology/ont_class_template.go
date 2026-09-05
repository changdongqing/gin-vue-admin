package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ClassTemplateRouter struct{}

// InitClassTemplateRouter 初始化分类模板路由（含编码规则端点）
func (r *ClassTemplateRouter) InitClassTemplateRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("classTemplate").Use(middleware.OperationRecord())
	read := priv.Group("classTemplate")
	{
		write.POST("createClassTemplate", classApi.CreateClassTemplate)   // 创建
		write.PUT("updateClassTemplate", classApi.UpdateClassTemplate)    // 更新（骨架 diff + 改父平移）
		write.PUT("disableClassTemplate", classApi.DisableClassTemplate)  // 弃用切换（预留）
		write.DELETE("deleteClassTemplate", classApi.DeleteClassTemplate) // 删除（有子拒绝）
	}
	{
		read.GET("findClassTemplate", classApi.FindClassTemplate)                 // 详情（含骨架）
		read.GET("getClassTemplateList", classApi.GetClassTemplateList)           // 扁平全量（组树用）
		read.GET("getClassTemplatePage", classApi.GetClassTemplatePage)           // 分页富化
		read.GET("getClassTemplateTreeRoots", classApi.GetClassTemplateTreeRoots) // 分类树下拉
		read.GET("getClassTemplateRefList", classApi.GetClassTemplateRefList)     // 骨架子表回填
		read.GET("getClassTemplateInherited", classApi.GetClassTemplateInherited) // 继承视图（预留）
		read.GET("previewClassificationCode", classApi.PreviewClassificationCode) // 编码预览
	}
	ruleWrite := priv.Group("classificationRule").Use(middleware.OperationRecord())
	ruleRead := priv.Group("classificationRule")
	{
		ruleWrite.PUT("saveClassificationRule", ruleApi.SaveClassificationRule) // 保存编码规则
		ruleRead.GET("findClassificationRule", ruleApi.FindClassificationRule)  // 查询编码规则
	}
}
