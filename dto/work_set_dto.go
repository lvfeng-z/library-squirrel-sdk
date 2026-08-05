package dto

import "github.com/lvfeng-z/library-squirrel-sdk/gen"

// WorkSetDTO 作品集数据传输对象（别名 gen.WorkSet，proto 单源）
type WorkSetDTO = gen.WorkSet

// WorkOrderEntry 作品在作品集内的原站排序条目（别名 gen.WorkOrderEntry，proto 单源）
type WorkOrderEntry = gen.WorkOrderEntry

// WorkOrderQuerier 可选能力接口：插件实现此接口以提供作品集内作品的原站顺序
// 未实现此接口的插件，主程序查询原站序时得到空响应（site_sort_order 保持空，仅本地序生效）
type WorkOrderQuerier interface {
	// QueryWorkSetOrder 返回作品集内作品的原站全序；空切片=插件不掌握
	QueryWorkSetOrder(siteId int64, siteWorkSetId string) ([]*WorkOrderEntry, error)
}

// WorkSetRelationEntry 作品集父集关系条目：本作品集的某个父集 + 在该父集下的原站序（别名 gen.WorkSetRelationEntry，proto 单源）
type WorkSetRelationEntry = gen.WorkSetRelationEntry

// WorkSetRelationQuerier 可选能力接口：插件实现此接口以提供作品集的父集关系 + 本集在各父集下的原站序。
// 与 WorkOrderQuerier 的区别：WorkOrderQuerier 提供「作品集内作品」的序；本接口提供「作品集间父子关系」+ 子集序。
// 未实现此接口的插件，主程序查询父集关系时得到空响应（site_sort_order 保持空，作品集层级不生效，扁平 workset 照常工作）。
type WorkSetRelationQuerier interface {
	// QueryWorkSetRelations 返回本作品集的父集 + 在各父集下的原站序；空切片=无父/插件不掌握
	QueryWorkSetRelations(siteId int64, siteWorkSetId string) ([]*WorkSetRelationEntry, error)
}
