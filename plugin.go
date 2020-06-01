package nested

import (
	"github.com/jinzhu/gorm"
)

const (
	tagName             = "nested"
	callbackNameCreate  = "nested:create"
	callbackNameUpdate  = "nested:update"
	callbackNameDelete  = "nested:delete"
	settingIgnoreUpdate = "nested:ignore_update"
	settingIgnoreDelete = "nested:ignore_delete"
)

// Plugin gorm nested set plugin
type Plugin struct {
	db            *gorm.DB
	treeLeftName  string
	treeRightName string
	treeLevelName string
}

// Register registers nested set plugin
func Register(db *gorm.DB) (Plugin, error) {
	p := Plugin{db: db}

	p.enableCallbacks()

	return p, nil
}

func (p *Plugin) enableCallbacks() {
	callback := p.db.Callback()
	callback.Create().After("gorm:after_create").Register(callbackNameCreate, p.createCallback)
	// todo: 禁用 update hook, 1. 移动用户节点、彩票返奖率设置强关联; 2. 目前实现每次用户记录更新重新计算 left, right 值并且错误
	//callback.Update().After("gorm:after_update").Register(callbackNameUpdate, p.updateCallback)
	callback.Delete().After("gorm:after_delete").Register(callbackNameDelete, p.deleteCallback)
}

// Interface must be implemented by the gorm model
type Interface interface {
	GetParentID() interface{}
	GetParent() Interface
}
