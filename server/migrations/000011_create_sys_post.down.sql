-- 岗位管理回滚：删除员工-岗位关联与岗位表
-- 字典 post_type 属运营数据，不在 down 中回滚

DROP TABLE IF EXISTS sys_user_post;
DROP TABLE IF EXISTS sys_posts;
