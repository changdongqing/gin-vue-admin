-- ============================================================
-- 岗位管理：sys_posts 岗位表 + sys_user_post 员工-岗位关联表
-- 参考 JeeSite js_sys_post / js_user_post（关联改 ID，岗位不挂公司/部门）
-- ============================================================

CREATE TABLE IF NOT EXISTS sys_posts (
    id          bigserial     PRIMARY KEY,
    created_at  timestamptz   NOT NULL DEFAULT now(),
    updated_at  timestamptz   NOT NULL DEFAULT now(),
    deleted_at  timestamptz,
    post_code   varchar(64)   NOT NULL,
    post_name   varchar(64)   NOT NULL,
    post_type   varchar(32)   NOT NULL DEFAULT '',
    sort        int           NOT NULL DEFAULT 0,
    status      smallint      NOT NULL DEFAULT 1,
    remarks     varchar(255)  NOT NULL DEFAULT ''
);
COMMENT ON TABLE sys_posts IS '岗位表（参考 JeeSite js_sys_post，独立模块，不涉数据权限）';
COMMENT ON COLUMN sys_posts.post_code IS '岗位编码（全局唯一，如 PM/HR）';
COMMENT ON COLUMN sys_posts.post_name IS '岗位名称（如 项目经理）';
COMMENT ON COLUMN sys_posts.post_type IS '岗位类型（字典 post_type 的 value）';
COMMENT ON COLUMN sys_posts.sort IS '排序（越小越靠前）';
COMMENT ON COLUMN sys_posts.status IS '状态 1启用 2停用';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_posts_code ON sys_posts (post_code);
CREATE INDEX IF NOT EXISTS idx_sys_posts_type ON sys_posts (post_type);
CREATE INDEX IF NOT EXISTS idx_sys_posts_deleted_at ON sys_posts (deleted_at);

CREATE TABLE IF NOT EXISTS sys_user_post (
    user_id     bigint   NOT NULL,
    post_id     bigint   NOT NULL,
    PRIMARY KEY (user_id, post_id)
);
COMMENT ON TABLE sys_user_post IS '员工-岗位关联表（N:N 双向绑定，硬删，无软删除）';
CREATE INDEX IF NOT EXISTS idx_sys_user_post_post ON sys_user_post (post_id);

-- ============================================================
-- 岗位类型字典种子（幂等）
-- ============================================================
INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '岗位类型', 'post_type', true, '岗位管理-岗位类型（参考 JeeSite sys_post_type）', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'post_type');

INSERT INTO sys_dictionary_details (label, value, sort, status, sys_dictionary_id, created_at, updated_at)
SELECT v.label, v.value, v.sort, true, d.id, now(), now()
FROM sys_dictionaries d,
     (VALUES ('高层岗位', 'high', 1), ('中层岗位', 'middle', 2),
             ('基层岗位', 'primary', 3), ('普通岗位', 'common', 4)) AS v(label, value, sort)
WHERE d.type = 'post_type'
  AND NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details x
    WHERE x.sys_dictionary_id = d.id AND x.value = v.value
  );
