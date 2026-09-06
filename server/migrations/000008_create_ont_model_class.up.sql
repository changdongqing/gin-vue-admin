-- ============================================================
-- 迁移: 000008_create_ont_model_class.up.sql
-- 目的: 本体建模·02 本体类建模
--       ont_model_classes（类实体）+ ont_model_datatype_properties（数据属性）
--       + ont_model_object_properties（对象属性）+ ont_model_subclassofs（层级，预留）
--       + 字典 ont_model_xsd_type
-- 注: ont_class_hierarchies（治理域镜像表）已随迁移 000003 建表预留，不在此重复
-- ============================================================

-- 本体类实体
CREATE TABLE IF NOT EXISTS ont_model_classes (
    id                   bigserial     PRIMARY KEY,
    project_id           bigint        NOT NULL,
    class_iri            varchar(255)  NOT NULL,
    local_name           varchar(128)  NOT NULL,
    label                varchar(128)  NOT NULL DEFAULT '',
    label_cn             varchar(128)  NOT NULL DEFAULT '',
    description          varchar(512)  NOT NULL DEFAULT '',
    template_code        varchar(64)   NOT NULL DEFAULT '',
    classification_code  varchar(64)   NOT NULL DEFAULT '',
    icon                 varchar(100)  NOT NULL DEFAULT '',
    color                varchar(50)   NOT NULL DEFAULT '',
    sort_order           int           NOT NULL DEFAULT 0,
    is_instantiable      smallint      NOT NULL DEFAULT 0,
    table_name           varchar(128)  NOT NULL DEFAULT '',
    created_by           varchar(64)   NOT NULL DEFAULT '',
    updated_by           varchar(64)   NOT NULL DEFAULT '',
    created_at           timestamptz   NOT NULL DEFAULT now(),
    updated_at           timestamptz   NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
COMMENT ON TABLE ont_model_classes IS '本体类实体（owl:Class，模板实例化/空白创建）';
COMMENT ON COLUMN ont_model_classes.project_id IS '所属项目ID → ont_model_projects.id';
COMMENT ON COLUMN ont_model_classes.class_iri IS '类IRI（namespace_base+local_name，项目内唯一）';
COMMENT ON COLUMN ont_model_classes.local_name IS '本地名（创建后不可改）';
COMMENT ON COLUMN ont_model_classes.template_code IS '溯源：来源分类模板编码（空=空白类）';
COMMENT ON COLUMN ont_model_classes.classification_code IS '溯源：来源分类编码（空=空白类）';
COMMENT ON COLUMN ont_model_classes.icon IS '图标（可继承分类模板）';
COMMENT ON COLUMN ont_model_classes.color IS '颜色hex（可继承分类模板）';
COMMENT ON COLUMN ont_model_classes.is_instantiable IS '是否实例化为主数据个体 0/1（外部模块关联仅绑1）';
COMMENT ON COLUMN ont_model_classes.table_name IS '预留：对象域主数据物理表名（业务表映射已迁出至外部模块关联，本期不写）';
COMMENT ON COLUMN ont_model_classes.created_by IS '创建人';
COMMENT ON COLUMN ont_model_classes.updated_by IS '更新人';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_model_class_iri
    ON ont_model_classes (project_id, class_iri) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_class_project
    ON ont_model_classes (project_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_class_template
    ON ont_model_classes (template_code) WHERE deleted_at IS NULL;

-- 数据属性（owl:DatatypeProperty）
CREATE TABLE IF NOT EXISTS ont_model_datatype_properties (
    id              bigserial     PRIMARY KEY,
    project_id      bigint        NOT NULL,
    class_id        bigint        NOT NULL,
    property_iri    varchar(255)  NOT NULL,
    local_name      varchar(128)  NOT NULL,
    label           varchar(128)  NOT NULL DEFAULT '',
    template_code   varchar(64)   NOT NULL DEFAULT '',
    xsd_type        varchar(64)   NOT NULL DEFAULT 'xsd:string',
    unit_ref        varchar(255)  NOT NULL DEFAULT '',
    enum_values     text          NOT NULL DEFAULT '',
    min_cardinality int           NOT NULL DEFAULT 0,
    max_cardinality int           NOT NULL DEFAULT -1,
    is_identifier   smallint      NOT NULL DEFAULT 0,
    sort_order      int           NOT NULL DEFAULT 0,
    created_by      varchar(64)   NOT NULL DEFAULT '',
    updated_by      varchar(64)   NOT NULL DEFAULT '',
    created_at      timestamptz   NOT NULL DEFAULT now(),
    updated_at      timestamptz   NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);
COMMENT ON TABLE ont_model_datatype_properties IS '本体数据属性（owl:DatatypeProperty，按类独立副本）';
COMMENT ON COLUMN ont_model_datatype_properties.property_iri IS '属性IRI（方案B：{ns}{classLocalName}_{localName}）';
COMMENT ON COLUMN ont_model_datatype_properties.template_code IS '溯源：属性模板编码（空=手建）';
COMMENT ON COLUMN ont_model_datatype_properties.xsd_type IS 'XSD类型 xsd:string/xsd:integer/xsd:decimal/xsd:boolean/xsd:datetime';
COMMENT ON COLUMN ont_model_datatype_properties.unit_ref IS '单位引用（QUDT IRI，仅数值型可绑）';
COMMENT ON COLUMN ont_model_datatype_properties.enum_values IS '枚举值（JSON数组字符串 → owl:oneOf）';
COMMENT ON COLUMN ont_model_datatype_properties.max_cardinality IS '最大基数 -1=无限制';
COMMENT ON COLUMN ont_model_datatype_properties.created_by IS '创建人';
COMMENT ON COLUMN ont_model_datatype_properties.updated_by IS '更新人';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_model_dt_prop_iri
    ON ont_model_datatype_properties (project_id, property_iri) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_dt_prop_class
    ON ont_model_datatype_properties (class_id) WHERE deleted_at IS NULL;

-- 对象属性（owl:ObjectProperty）
CREATE TABLE IF NOT EXISTS ont_model_object_properties (
    id              bigserial     PRIMARY KEY,
    project_id      bigint        NOT NULL,
    domain_class_id bigint        NOT NULL,
    range_class_id  bigint,
    property_iri    varchar(255)  NOT NULL,
    local_name      varchar(128)  NOT NULL,
    label           varchar(128)  NOT NULL DEFAULT '',
    template_code   varchar(64)   NOT NULL DEFAULT '',
    min_cardinality int           NOT NULL DEFAULT 0,
    max_cardinality int           NOT NULL DEFAULT -1,
    inverse_of      bigint,
    sort_order      int           NOT NULL DEFAULT 0,
    created_by      varchar(64)   NOT NULL DEFAULT '',
    updated_by      varchar(64)   NOT NULL DEFAULT '',
    created_at      timestamptz   NOT NULL DEFAULT now(),
    updated_at      timestamptz   NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);
COMMENT ON TABLE ont_model_object_properties IS '本体对象属性（owl:ObjectProperty，domain=所属类）';
COMMENT ON COLUMN ont_model_object_properties.domain_class_id IS '域类ID（=所属类，自动填充）';
COMMENT ON COLUMN ont_model_object_properties.range_class_id IS '值域类ID（可空，模板实例化骨架留空待补）';
COMMENT ON COLUMN ont_model_object_properties.inverse_of IS '反向属性ID（owl:inverseOf 互指）';
COMMENT ON COLUMN ont_model_object_properties.created_by IS '创建人';
COMMENT ON COLUMN ont_model_object_properties.updated_by IS '更新人';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_model_obj_prop_iri
    ON ont_model_object_properties (project_id, property_iri) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_obj_prop_domain
    ON ont_model_object_properties (domain_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_obj_prop_range
    ON ont_model_object_properties (range_class_id) WHERE deleted_at IS NULL;

-- 类层级关系（建模域权威源；本期建表预留，层级编辑/镜像回推为后续需求）
CREATE TABLE IF NOT EXISTS ont_model_subclassofs (
    id                   bigserial     PRIMARY KEY,
    project_id           bigint        NOT NULL,
    child_class_id       bigint        NOT NULL,
    parent_class_id      bigint        NOT NULL,
    source_template_code varchar(64)   NOT NULL DEFAULT '',
    sync_status          smallint      NOT NULL DEFAULT 0,
    sync_time            timestamptz,
    retry_count          int           NOT NULL DEFAULT 0,
    created_by           varchar(64)   NOT NULL DEFAULT '',
    updated_by           varchar(64)   NOT NULL DEFAULT '',
    created_at           timestamptz   NOT NULL DEFAULT now(),
    updated_at           timestamptz   NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
COMMENT ON TABLE ont_model_subclassofs IS 'subClassOf 类层级关系（建模域权威源；本期仅建表供删除守卫与详情聚合，编辑功能后续交付）';
COMMENT ON COLUMN ont_model_subclassofs.sync_status IS '镜像回推状态 0待同步/1已同步/2已失效';
COMMENT ON COLUMN ont_model_subclassofs.created_by IS '创建人';
COMMENT ON COLUMN ont_model_subclassofs.updated_by IS '更新人';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_model_subclassof
    ON ont_model_subclassofs (project_id, child_class_id, parent_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_model_subclassof_parent
    ON ont_model_subclassofs (parent_class_id) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------
-- 字典种子：ont_model_xsd_type
-- ------------------------------------------------------------
INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体XSD类型', 'ont_model_xsd_type', true, '本体数据属性值域XSD类型', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_model_xsd_type');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT x.label, x.value, true, x.sort, (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_xsd_type'), now(), now()
FROM (VALUES
    ('字符串',   'xsd:string',   1),
    ('整数',     'xsd:integer',  2),
    ('小数',     'xsd:decimal',  3),
    ('布尔',     'xsd:boolean',  4),
    ('日期时间', 'xsd:datetime', 5)
) AS x(label, value, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_xsd_type')
      AND d.value = x.value);
