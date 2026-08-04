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
