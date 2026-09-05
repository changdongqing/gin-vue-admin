-- ============================================================
-- 迁移: 000002_create_ont_property_template.up.sql
-- 目的: 本体属性模板表（数据属性 + 对象属性同构）+ 字典种子 + 内置模板种子
-- 说明: 主键 uint 自增（GVA_MODEL，PG 序列）；软删除 deleted_at；
--       唯一索引为部分索引 WHERE deleted_at IS NULL（逻辑删后允许编码复用）
-- ============================================================

CREATE TABLE IF NOT EXISTS ont_property_templates (
    id                  bigserial       PRIMARY KEY,
    template_code       varchar(64)     NOT NULL,
    kind                varchar(16)     NOT NULL DEFAULT 'datatype',
    label               varchar(128)    NOT NULL,
    alias               varchar(512)    NOT NULL DEFAULT '',
    description         varchar(512)    NOT NULL DEFAULT '',
    category            varchar(64)     NOT NULL DEFAULT '',
    type                varchar(32)     NOT NULL DEFAULT '',
    is_identifier       smallint        NOT NULL DEFAULT 0,
    unit_ref            varchar(255)    NOT NULL DEFAULT '',
    "values"            varchar(1000)   NOT NULL DEFAULT '',
    default_cardinality varchar(32)     NOT NULL DEFAULT '',
    source              varchar(16)     NOT NULL DEFAULT 'custom',
    deprecated          smallint        NOT NULL DEFAULT 0,
    status              smallint        NOT NULL DEFAULT 0,
    created_by          varchar(64)     NOT NULL DEFAULT '',
    updated_by          varchar(64)     NOT NULL DEFAULT '',
    created_at          timestamptz     NOT NULL DEFAULT now(),
    updated_at          timestamptz     NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);

COMMENT ON TABLE ont_property_templates IS '本体属性模板表（数据属性 + 对象属性同构）';
COMMENT ON COLUMN ont_property_templates.template_code IS '模板编码（唯一业务标识，如 name/contains/price）';
COMMENT ON COLUMN ont_property_templates.kind IS '属性类型: datatype(数据属性)/object(对象属性)';
COMMENT ON COLUMN ont_property_templates.label IS '显示名';
COMMENT ON COLUMN ont_property_templates.alias IS '同义别名（逗号分隔），供搜索命中与治理合并留痕';
COMMENT ON COLUMN ont_property_templates.description IS '业务说明';
COMMENT ON COLUMN ont_property_templates.category IS '分组: basic/contact/monetary/temporal/status/containment/attribution/rated/monitoring/running/management/operation';
COMMENT ON COLUMN ont_property_templates.type IS '数据类型(datatype 专属): string/integer/decimal/boolean/datetime';
COMMENT ON COLUMN ont_property_templates.is_identifier IS '是否标识(datatype 专属): 0=否 1=是';
COMMENT ON COLUMN ont_property_templates.unit_ref IS '预设 QUDT 单位 IRI（与单位注册表联动）';
COMMENT ON COLUMN ont_property_templates."values" IS '枚举值(逗号分隔, datatype 专属)';
COMMENT ON COLUMN ont_property_templates.default_cardinality IS '默认基数(object 专属): one-to-many/many-to-one/one-to-one/many-to-many';
COMMENT ON COLUMN ont_property_templates.source IS '来源: builtin(内置)/custom(自定义)';
COMMENT ON COLUMN ont_property_templates.deprecated IS '是否弃用: 0=否 1=是';
COMMENT ON COLUMN ont_property_templates.status IS '业务启停: 0=正常 1=停用';
COMMENT ON COLUMN ont_property_templates.created_by IS '创建人（用户名，API 层填充）';
COMMENT ON COLUMN ont_property_templates.updated_by IS '更新人（用户名，API 层填充）';

CREATE INDEX IF NOT EXISTS idx_ont_prop_tpl_kind ON ont_property_templates (kind) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_prop_tpl_category ON ont_property_templates (category) WHERE deleted_at IS NULL;
-- 唯一业务编码：软删除后允许复用
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_prop_tpl_code
    ON ont_property_templates (template_code) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------
-- 内置属性模板种子（source='builtin'，治理侧不可编辑/删除，可弃用）
-- ------------------------------------------------------------
INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'name', 'datatype', '名称', 'basic', 'string', 1, '', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'name' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'description', 'datatype', '描述', 'basic', 'string', 0, '', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'description' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'email', 'datatype', '邮箱', 'contact', 'string', 0, '', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'email' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'phone', 'datatype', '电话', 'contact', 'string', 0, '', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'phone' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'price', 'datatype', '价格', 'monetary', 'decimal', 0, 'http://qudt.org/vocab/unit/USD', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'price' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'amount', 'datatype', '金额', 'monetary', 'decimal', 0, 'http://qudt.org/vocab/unit/USD', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'amount' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'createdDate', 'datatype', '创建日期', 'temporal', 'datetime', 0, '', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'createdDate' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'isActive', 'datatype', '是否启用', 'status', 'boolean', 0, '', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'isActive' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'length', 'datatype', '长度', 'basic', 'decimal', 0, 'http://qudt.org/vocab/unit/M', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'length' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, default_cardinality, source)
SELECT 'contains', 'object', '包含', 'containment', '', 0, '', 'one-to-many', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'contains' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, default_cardinality, source)
SELECT 'belongsTo', 'object', '属于', 'attribution', '', 0, '', 'many-to-one', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'belongsTo' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, default_cardinality, source)
SELECT 'hasPart', 'object', '拥有部件', 'containment', '', 0, '', 'one-to-many', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'hasPart' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, default_cardinality, source)
SELECT 'references', 'object', '引用', 'attribution', '', 0, '', 'many-to-one', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'references' AND deleted_at IS NULL);

-- ------------------------------------------------------------
-- 字典种子：ont_property_kind / ont_source（前端 getDict/gvaDict 渲染）
-- ------------------------------------------------------------
INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体属性类型', 'ont_property_kind', true, '数据属性/对象属性', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_property_kind');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT '数据属性', 'datatype', true, 1, (SELECT id FROM sys_dictionaries WHERE type = 'ont_property_kind'), now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_property_kind')
      AND d.value = 'datatype');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT '对象属性', 'object', true, 2, (SELECT id FROM sys_dictionaries WHERE type = 'ont_property_kind'), now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_property_kind')
      AND d.value = 'object');

INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体来源', 'ont_source', true, '内置/自定义', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_source');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT '内置', 'builtin', true, 1, (SELECT id FROM sys_dictionaries WHERE type = 'ont_source'), now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_source')
      AND d.value = 'builtin');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT '自定义', 'custom', true, 2, (SELECT id FROM sys_dictionaries WHERE type = 'ont_source'), now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_source')
      AND d.value = 'custom');
