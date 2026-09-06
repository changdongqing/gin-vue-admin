-- ============================================================
-- 迁移: 000007_create_ont_model_project.up.sql
-- 目的: 本体建模·01 本体项目管理（建模域顶层容器）
--       ont_model_projects（项目）+ ont_model_prefixes（IRI前缀注册）
--       + 字典 ont_model_project_status / ont_model_format / ont_model_strategy
-- ============================================================

-- 本体项目（建模域顶层容器）
CREATE TABLE IF NOT EXISTS ont_model_projects (
    id                     bigserial     PRIMARY KEY,
    project_code           varchar(64)   NOT NULL,
    name                   varchar(128)  NOT NULL,
    description            varchar(512)  NOT NULL DEFAULT '',
    namespace_base         varchar(255)  NOT NULL,
    default_format         varchar(16)   NOT NULL DEFAULT 'TTL',
    serialization_strategy varchar(8)    NOT NULL DEFAULT 'B',
    status                 varchar(16)   NOT NULL DEFAULT 'draft',
    created_by             varchar(64)   NOT NULL DEFAULT '',
    updated_by             varchar(64)   NOT NULL DEFAULT '',
    created_at             timestamptz   NOT NULL DEFAULT now(),
    updated_at             timestamptz   NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);
COMMENT ON TABLE ont_model_projects IS '本体项目（建模域顶层容器：命名空间/前缀/序列化策略）';
COMMENT ON COLUMN ont_model_projects.project_code IS '项目编码（全局唯一，如 fire-equipment）';
COMMENT ON COLUMN ont_model_projects.name IS '项目名称';
COMMENT ON COLUMN ont_model_projects.description IS '描述';
COMMENT ON COLUMN ont_model_projects.namespace_base IS 'IRI命名空间基址（classIri = namespace_base + localName）';
COMMENT ON COLUMN ont_model_projects.default_format IS '默认序列化格式 TTL/OWL_XML';
COMMENT ON COLUMN ont_model_projects.serialization_strategy IS '序列化策略 B=带前缀独立副本(默认) A=共享属性+多domain(预留禁用)';
COMMENT ON COLUMN ont_model_projects.status IS '项目状态 draft草稿/active活跃/archived归档(只读)';
COMMENT ON COLUMN ont_model_projects.created_by IS '创建人';
COMMENT ON COLUMN ont_model_projects.updated_by IS '更新人';

-- projectCode 全局唯一（软删后允许复用）
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_model_project_code
    ON ont_model_projects (project_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_project_status
    ON ont_model_projects (status) WHERE deleted_at IS NULL;

-- IRI 前缀注册（项目级）
CREATE TABLE IF NOT EXISTS ont_model_prefixes (
    id          bigserial     PRIMARY KEY,
    project_id  bigint        NOT NULL,
    prefix      varchar(64)   NOT NULL,
    namespace   varchar(255)  NOT NULL,
    is_default  smallint      NOT NULL DEFAULT 0,
    created_by  varchar(64)   NOT NULL DEFAULT '',
    updated_by  varchar(64)   NOT NULL DEFAULT '',
    created_at  timestamptz   NOT NULL DEFAULT now(),
    updated_at  timestamptz   NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
COMMENT ON TABLE ont_model_prefixes IS '本体项目 IRI 前缀注册（序列化 @prefix 来源）';
COMMENT ON COLUMN ont_model_prefixes.project_id IS '所属项目ID → ont_model_projects.id';
COMMENT ON COLUMN ont_model_prefixes.prefix IS '前缀名（NCName，如 ex/qudt/fire）';
COMMENT ON COLUMN ont_model_prefixes.namespace IS '命名空间 URI';
COMMENT ON COLUMN ont_model_prefixes.is_default IS '是否项目默认前缀 0/1（同项目至多一个1）';
COMMENT ON COLUMN ont_model_prefixes.created_by IS '创建人';
COMMENT ON COLUMN ont_model_prefixes.updated_by IS '更新人';

-- 同项目 prefix 唯一（软删后允许复用）
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_model_prefix
    ON ont_model_prefixes (project_id, prefix) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_prefix_project
    ON ont_model_prefixes (project_id) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------
-- 字典种子：ont_model_project_status / ont_model_format / ont_model_strategy
-- ------------------------------------------------------------
INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体项目状态', 'ont_model_project_status', true, '本体建模项目状态机', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_model_project_status');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT x.label, x.value, true, x.sort, (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_project_status'), now(), now()
FROM (VALUES
    ('草稿', 'draft',    1),
    ('活跃', 'active',   2),
    ('归档', 'archived', 3)
) AS x(label, value, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_project_status')
      AND d.value = x.value);

INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体序列化格式', 'ont_model_format', true, '本体项目默认序列化格式', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_model_format');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT x.label, x.value, true, x.sort, (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_format'), now(), now()
FROM (VALUES
    ('Turtle',  'TTL',     1),
    ('OWL XML', 'OWL_XML', 2)
) AS x(label, value, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_format')
      AND d.value = x.value);

INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体序列化策略', 'ont_model_strategy', true, '本体属性组织方案（A预留禁用）', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_model_strategy');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT x.label, x.value, true, x.sort, (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_strategy'), now(), now()
FROM (VALUES
    ('方案B·独立副本',     'B', 1),
    ('方案A·共享属性（预留）', 'A', 2)
) AS x(label, value, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_strategy')
      AND d.value = x.value);
