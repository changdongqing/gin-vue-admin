package system

import (
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"gorm.io/gorm"
)

type PostService struct{}

var PostServiceApp = new(PostService)

// dedupeUint 有序去重
func dedupeUint(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// CreatePost 新建岗位（编码全局唯一，含软删行——与 DDL 全局唯一索引口径一致）
func (postService *PostService) CreatePost(post *system.SysPost) error {
	var count int64
	global.GVA_DB.Unscoped().Model(&system.SysPost{}).Where("post_code = ?", post.PostCode).Count(&count)
	if count > 0 {
		return errors.New("岗位编码已存在")
	}
	return global.GVA_DB.Create(post).Error
}

// DeletePost 删除岗位（有关联员工禁止删除；无关联软删除）
func (postService *PostService) DeletePost(id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&system.SysUserPost{}).Where("post_id = ?", id).Count(&count)
		if count > 0 {
			return fmt.Errorf("该岗位已关联 %d 名员工，请先解绑后再删除", count)
		}
		return tx.Delete(&system.SysPost{}, "id = ?", id).Error
	})
}

// UpdatePost 更新岗位（编码唯一性校验排除自身）
func (postService *PostService) UpdatePost(post *system.SysPost) error {
	var count int64
	global.GVA_DB.Unscoped().Model(&system.SysPost{}).
		Where("post_code = ? AND id <> ?", post.PostCode, post.ID).Count(&count)
	if count > 0 {
		return errors.New("岗位编码已存在")
	}
	return global.GVA_DB.Save(post).Error
}

// GetPost 岗位详情
func (postService *PostService) GetPost(id uint) (post system.SysPost, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&post).Error
	return
}

// GetPostList 分页列表（keyword 匹配编码/名称 + 类型 + 状态筛选，携带关联员工数）
func (postService *PostService) GetPostList(info systemReq.SysPostSearch) (list []system.SysPostListItem, total int64, err error) {
	db := global.GVA_DB.Model(&system.SysPost{})
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("post_code LIKE ? OR post_name LIKE ?", kw, kw)
	}
	if info.PostType != "" {
		db = db.Where("post_type = ?", info.PostType)
	}
	if info.Status != 0 {
		db = db.Where("status = ?", info.Status)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Select("sys_posts.*, (SELECT COUNT(1) FROM sys_user_post up WHERE up.post_id = sys_posts.id) AS user_count").
		Order("sort, id").
		Scopes(info.Paginate()).
		Scan(&list).Error
	return
}

// GetPostListAll 全部启用岗位（不分页——PageInfo.Paginate 会将 pageSize 截为 ≤100，全量场景独立查询）
func (postService *PostService) GetPostListAll() (list []system.SysPost, err error) {
	err = global.GVA_DB.Where("status = ?", 1).Order("sort, id").Find(&list).Error
	return
}

// GetPostUsers 岗位已关联员工全量（非分页，供分配员工弹窗回显）
func (postService *PostService) GetPostUsers(postId uint) (list []system.PostUserBrief, err error) {
	err = global.GVA_DB.Table("sys_user_post").
		Select("sys_users.id AS id, sys_users.username AS user_name, sys_users.nick_name AS nick_name, sys_users.department_id AS department_id, sys_departments.name AS department_name").
		Joins("JOIN sys_users ON sys_users.id = sys_user_post.user_id AND sys_users.deleted_at IS NULL").
		Joins("LEFT JOIN sys_departments ON sys_departments.id = sys_users.department_id AND sys_departments.deleted_at IS NULL").
		Where("sys_user_post.post_id = ?", postId).
		Order("sys_users.id").
		Scan(&list).Error
	return
}

// SetPostUsers 岗位绑定员工（全量覆盖：事务内 diff，仅校验存在性，停用不拦截——解绑场景仍需可用）
func (postService *PostService) SetPostUsers(postId uint, userIds []uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var postCount int64
		tx.Unscoped().Model(&system.SysPost{}).Where("id = ?", postId).Count(&postCount)
		if postCount == 0 {
			return errors.New("岗位不存在")
		}
		return syncUserPost(tx, postId, userIds)
	})
}

// GetUserPosts 员工已关联岗位（非分页，含已挂的停用岗位——回显要如实）
func (postService *PostService) GetUserPosts(userId uint) (list []system.SysPost, err error) {
	err = global.GVA_DB.
		Joins("JOIN sys_user_post up ON up.post_id = sys_posts.id AND up.user_id = ?", userId).
		Order("sys_posts.sort, sys_posts.id").
		Find(&list).Error
	return
}

// SetUserPosts 员工绑定岗位（全量覆盖，与 SetPostUsers 对称）
func (postService *PostService) SetUserPosts(userId uint, postIds []uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var userCount int64
		tx.Unscoped().Model(&system.SysUser{}).Where("id = ?", userId).Count(&userCount)
		if userCount == 0 {
			return errors.New("用户不存在")
		}
		// 校验岗位均存在（未软删）
		ids := dedupeUint(postIds)
		if len(ids) > 0 {
			var postCount int64
			tx.Model(&system.SysPost{}).Where("id IN ?", ids).Count(&postCount)
			if int(postCount) != len(ids) {
				return errors.New("存在无效的岗位ID")
			}
		}
		var old []system.SysUserPost
		tx.Where("user_id = ?", userId).Find(&old)
		oldSet := make(map[uint]struct{}, len(old))
		for _, o := range old {
			oldSet[o.PostId] = struct{}{}
		}
		newSet := make(map[uint]struct{}, len(ids))
		for _, p := range ids {
			newSet[p] = struct{}{}
		}
		var toRemove, toAdd []uint
		for _, o := range old {
			if _, ok := newSet[o.PostId]; !ok {
				toRemove = append(toRemove, o.PostId)
			}
		}
		for _, p := range ids {
			if _, ok := oldSet[p]; !ok {
				toAdd = append(toAdd, p)
			}
		}
		if len(toRemove) > 0 {
			if err := tx.Where("user_id = ? AND post_id IN ?", userId, toRemove).Delete(&system.SysUserPost{}).Error; err != nil {
				return err
			}
		}
		if len(toAdd) > 0 {
			rows := make([]system.SysUserPost, 0, len(toAdd))
			for _, p := range toAdd {
				rows = append(rows, system.SysUserPost{UserId: userId, PostId: p})
			}
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// syncUserPost 以岗位方向做全量覆盖 diff（调用方已校验岗位存在性）
func syncUserPost(tx *gorm.DB, postId uint, userIds []uint) error {
	ids := dedupeUint(userIds)
	// 校验用户均存在且未软删
	if len(ids) > 0 {
		var userCount int64
		tx.Model(&system.SysUser{}).Where("id IN ?", ids).Count(&userCount)
		if int(userCount) != len(ids) {
			return errors.New("存在无效的用户ID")
		}
	}
	var old []system.SysUserPost
	tx.Where("post_id = ?", postId).Find(&old)
	oldSet := make(map[uint]struct{}, len(old))
	for _, o := range old {
		oldSet[o.UserId] = struct{}{}
	}
	newSet := make(map[uint]struct{}, len(ids))
	for _, u := range ids {
		newSet[u] = struct{}{}
	}
	var toRemove, toAdd []uint
	for _, o := range old {
		if _, ok := newSet[o.UserId]; !ok {
			toRemove = append(toRemove, o.UserId)
		}
	}
	for _, u := range ids {
		if _, ok := oldSet[u]; !ok {
			toAdd = append(toAdd, u)
		}
	}
	if len(toRemove) > 0 {
		if err := tx.Where("post_id = ? AND user_id IN ?", postId, toRemove).Delete(&system.SysUserPost{}).Error; err != nil {
			return err
		}
	}
	if len(toAdd) > 0 {
		rows := make([]system.SysUserPost, 0, len(toAdd))
		for _, u := range toAdd {
			rows = append(rows, system.SysUserPost{UserId: u, PostId: postId})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	return nil
}
