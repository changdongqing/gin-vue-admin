-- ============================================================
-- 迁移: 000005_create_ont_annotation_property.up.sql
-- 目的: 本体注释属性注册表（ont:xxx 集中声明，建模侧序列化驱动）
--       + 字典 ont_applies_to + 12 条内置声明种子
-- 说明: 本表无 builtin 保护（声明式元数据，治理员可调整内置项）
-- ============================================================

CREATE TABLE IF NOT EXISTS ont_annotation_properties (
    id          bigserial     PRIMARY KEY,
    local_name  varchar(64)   NOT NULL,
    label       varchar(128)  NOT NULL,
    range_xsd   varchar(64)   NOT NULL DEFAULT '',
    applies_to  varchar(32)   NOT NULL DEFAULT 'all',
    description varchar(512)  NOT NULL DEFAULT '',
    sort        int           NOT NULL DEFAULT 0,
    created_by  varchar(64)   NOT NULL DEFAULT '',
    updated_by  varchar(64)   NOT NULL DEFAULT '',
    created_at  timestamptz   NOT NULL DEFAULT now(),
    updated_at  timestamptz   NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
COMMENT ON TABLE ont_annotation_properties IS '本体注释属性注册表（ont:xxx 集中声明，建模侧序列化驱动）';
COMMENT ON COLUMN ont_annotation_properties.local_name IS '注释属性名（唯一，如 icon/unitRef）';
COMMENT ON COLUMN ont_annotation_properties.label IS '显示名';
COMMENT ON COLUMN ont_annotation_properties.range_xsd IS '值域XSD类型（如 xsd:string/xsd:boolean/xsd:anyURI）';
COMMENT ON COLUMN ont_annotation_properties.applies_to IS '作用对象: class/datatypeProperty/objectProperty/individual/all';
COMMENT ON COLUMN ont_annotation_properties.description IS '描述';
COMMENT ON COLUMN ont_annotation_properties.sort IS '排序';
COMMENT ON COLUMN ont_annotation_properties.created_by IS '创建人';
COMMENT ON COLUMN ont_annotation_properties.updated_by IS '更新人';

-- localName 唯一（软删后允许复用）
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_ap_name
    ON ont_annotation_properties (local_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_ap_applies ON ont_annotation_properties (applies_to) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------
-- 内置声明种子（12 条；无 builtin 保护，治理员可编辑/删除）
-- ------------------------------------------------------------
INSERT INTO ont_annotation_properties (local_name, label, range_xsd, applies_to, description, sort)
SELECT x.local_name, x.label, x.range_xsd, x.applies_to, x.descr, x.sort
FROM (VALUES
    ('icon',            '图标',     'xsd:string',  'class',            '',                    1),
    ('color',           '颜色',     'xsd:string',  'class',            '',                    2),
    ('unit',            '单位',     'xsd:anyURI',  'datatypeProperty', '',                    3),
    ('unitRef',         '单位引用', 'xsd:anyURI',  'datatypeProperty', '关键项勿删',           4),
    ('quantityKindRef', '量纲引用', 'xsd:anyURI',  'datatypeProperty', '',                    5),
    ('propertyType',    '属性类型', 'xsd:string',  'datatypeProperty', '',                    6),
    ('cardinality',     '基数',     'xsd:string',  'objectProperty',   '',                    7),
    ('isIdentifier',    '是否标识', 'xsd:boolean', 'datatypeProperty', '',                    8),
    ('enumValues',      '枚举值',   'xsd:string',  'datatypeProperty', '',                    9),
    ('fromEntityId',    '来源实体', 'xsd:string',  'objectProperty',   '',                    10),
    ('toEntityId',      '目标实体', 'xsd:string',  'objectProperty',   '',                    11),
    ('templateRef',     '模板引用', 'xsd:string',  'all',              '关键项勿删',           12)
) AS x(local_name, label, range_xsd, applies_to, descr, sort)
WHERE NOT EXISTS (SELECT 1 FROM ont_annotation_properties WHERE local_name = x.local_name AND deleted_at IS NULL);

-- ------------------------------------------------------------
-- 字典种子：ont_applies_to
-- ------------------------------------------------------------
INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体作用对象', 'ont_applies_to', true, '注释属性作用对象', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_applies_to');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT x.label, x.value, true, x.sort, (SELECT id FROM sys_dictionaries WHERE type = 'ont_applies_to'), now(), now()
FROM (VALUES
    ('类',       'class',            1),
    ('数据属性', 'datatypeProperty', 2),
    ('对象属性', 'objectProperty',   3),
    ('个体',     'individual',       4),
    ('全部',     'all',              5)
) AS x(label, value, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_applies_to')
      AND d.value = x.value);
