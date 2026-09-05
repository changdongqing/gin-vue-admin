-- ============================================================
-- 迁移: 000003_create_ont_class_template.up.sql
-- 目的: 分类模板四表（主表/骨架子表/编码规则/类层级镜像）+ rated 属性模板种子
--       + equipment 分类树种子 + 编码规则种子 + ont_ref_type 字典
-- 说明: 树表采用物化路径 parent_path（含父链不含自身，如 0,12,35,），对齐 sys_companies 维护算法
-- ============================================================

-- 1. 分类模板主表
CREATE TABLE IF NOT EXISTS ont_class_templates (
    id                  bigserial     PRIMARY KEY,
    template_code       varchar(64)   NOT NULL,
    classification_code varchar(64)   NOT NULL DEFAULT '',
    label               varchar(128)  NOT NULL,
    label_cn            varchar(128)  NOT NULL DEFAULT '',
    description         varchar(512)  NOT NULL DEFAULT '',
    tree_root           varchar(64)   NOT NULL DEFAULT '',
    parent_id           bigint        NOT NULL DEFAULT 0,
    parent_path         varchar(767)  NOT NULL DEFAULT '0,',
    tree_level          int           NOT NULL DEFAULT 0,
    sort                int           NOT NULL DEFAULT 0,
    icon                varchar(100)  NOT NULL DEFAULT '',
    color               varchar(50)   NOT NULL DEFAULT '',
    inherit_appearance  smallint      NOT NULL DEFAULT 1,
    source              varchar(16)   NOT NULL DEFAULT 'custom',
    source_ref          varchar(64)   NOT NULL DEFAULT '',
    deprecated          smallint      NOT NULL DEFAULT 0,
    status              smallint      NOT NULL DEFAULT 0,
    created_by          varchar(64)   NOT NULL DEFAULT '',
    updated_by          varchar(64)   NOT NULL DEFAULT '',
    created_at          timestamptz   NOT NULL DEFAULT now(),
    updated_at          timestamptz   NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);
COMMENT ON TABLE ont_class_templates IS '本体分类模板表（外观+结构骨架+父子继承+编码，树表）';
COMMENT ON COLUMN ont_class_templates.template_code IS '模板编码（唯一，如 pump/spray-pump）';
COMMENT ON COLUMN ont_class_templates.classification_code IS '规范分类编码（如 30-01-01，tree_root 范围内唯一）';
COMMENT ON COLUMN ont_class_templates.label IS '显示名（如 喷淋泵）';
COMMENT ON COLUMN ont_class_templates.label_cn IS '中文名';
COMMENT ON COLUMN ont_class_templates.description IS '描述';
COMMENT ON COLUMN ont_class_templates.tree_root IS '所属分类树标识（如 equipment，支持多棵树）';
COMMENT ON COLUMN ont_class_templates.parent_id IS '父分类模板ID，根=0';
COMMENT ON COLUMN ont_class_templates.parent_path IS '物化路径（含父链不含自身，如 0,12,35,）';
COMMENT ON COLUMN ont_class_templates.tree_level IS '层级（根=0，上限 3=四级）';
COMMENT ON COLUMN ont_class_templates.sort IS '当前层级排序号';
COMMENT ON COLUMN ont_class_templates.icon IS '图标（emoji 或图标类名，外观）';
COMMENT ON COLUMN ont_class_templates.color IS '颜色 hex（外观）';
COMMENT ON COLUMN ont_class_templates.inherit_appearance IS '是否继承父外观: 0=否 1=是';
COMMENT ON COLUMN ont_class_templates.source IS '来源: builtin/custom';
COMMENT ON COLUMN ont_class_templates.source_ref IS '来源本体标识（如 brick，参考本体导入时写入）';
COMMENT ON COLUMN ont_class_templates.deprecated IS '是否弃用: 0/1';
COMMENT ON COLUMN ont_class_templates.status IS '业务启停: 0=正常 1=停用';
COMMENT ON COLUMN ont_class_templates.created_by IS '创建人（用户名，API 层填充）';
COMMENT ON COLUMN ont_class_templates.updated_by IS '更新人（用户名，API 层填充）';

-- 编码唯一（软删后允许复用）；分类编码 tree_root 范围内唯一（空编码不参与）
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_class_tpl_code
    ON ont_class_templates (template_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_class_tpl_cc
    ON ont_class_templates (tree_root, classification_code)
    WHERE deleted_at IS NULL AND classification_code <> '';
CREATE INDEX IF NOT EXISTS idx_ont_class_tpl_parent ON ont_class_templates (parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_class_tpl_root   ON ont_class_templates (tree_root) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_class_tpl_path   ON ont_class_templates (parent_path) WHERE deleted_at IS NULL;

-- 2. 结构骨架子表
CREATE TABLE IF NOT EXISTS ont_class_template_refs (
    id                     bigserial    PRIMARY KEY,
    class_template_id      bigint       NOT NULL,
    property_template_code varchar(64)  NOT NULL,
    ref_type               varchar(16)  NOT NULL DEFAULT 'property',
    sort_order             int          NOT NULL DEFAULT 0,
    created_at             timestamptz  NOT NULL DEFAULT now(),
    updated_at             timestamptz  NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);
COMMENT ON TABLE ont_class_template_refs IS '本体分类模板结构骨架引用表（主子表）';
COMMENT ON COLUMN ont_class_template_refs.class_template_id IS '所属分类模板ID';
COMMENT ON COLUMN ont_class_template_refs.property_template_code IS '引用的属性模板 template_code（业务软引用）';
COMMENT ON COLUMN ont_class_template_refs.ref_type IS '引用类型: property(属性)/relationship(关系)';
COMMENT ON COLUMN ont_class_template_refs.sort_order IS '注入顺序';
-- 同一主表内引用去重（软删后可重建）
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_class_tpl_ref
    ON ont_class_template_refs (class_template_id, property_template_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_class_tpl_ref_tid ON ont_class_template_refs (class_template_id);
-- 反查：属性模板被哪些骨架引用（删除属性模板前校验，01 篇 TODO 的落地）
CREATE INDEX IF NOT EXISTS idx_ont_class_tpl_ref_code ON ont_class_template_refs (property_template_code);

-- 3. 分类编码规则（每棵树一条）
CREATE TABLE IF NOT EXISTS ont_classification_rules (
    id           bigserial    PRIMARY KEY,
    tree_root    varchar(64)  NOT NULL,
    separator    varchar(4)   NOT NULL DEFAULT '-',
    level_digits int          NOT NULL DEFAULT 2,
    base_number  int          NOT NULL DEFAULT 30,
    zero_pad     smallint     NOT NULL DEFAULT 1,
    description  varchar(512) NOT NULL DEFAULT '',
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
COMMENT ON TABLE ont_classification_rules IS '本体分类编码规则表（FR-8）';
COMMENT ON COLUMN ont_classification_rules.tree_root IS '所属分类树标识（唯一）';
COMMENT ON COLUMN ont_classification_rules.separator IS '分隔符（默认 -）';
COMMENT ON COLUMN ont_classification_rules.level_digits IS '每级位数（默认 2）';
COMMENT ON COLUMN ont_classification_rules.base_number IS '根级编码基数（如 30）';
COMMENT ON COLUMN ont_classification_rules.zero_pad IS '是否零填充: 0/1（默认 1）';
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_cr_root
    ON ont_classification_rules (tree_root) WHERE deleted_at IS NULL;

-- 4. 类层级镜像（本期建表预留，不实现同步；建模侧为权威）
CREATE TABLE IF NOT EXISTS ont_class_hierarchies (
    id                   bigserial    PRIMARY KEY,
    child_class_iri      varchar(255) NOT NULL,
    parent_class_iri     varchar(255) NOT NULL,
    source_template_code varchar(64)  NOT NULL DEFAULT '',
    tree_root            varchar(64)  NOT NULL DEFAULT '',
    sync_status          smallint     NOT NULL DEFAULT 0,
    sync_time            timestamptz,
    created_at           timestamptz  NOT NULL DEFAULT now(),
    updated_at           timestamptz  NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);
COMMENT ON TABLE ont_class_hierarchies IS '本体类分类树镜像表（建模侧 subClassOf 镜像，本期建表预留）';
COMMENT ON COLUMN ont_class_hierarchies.sync_status IS '同步状态: 0=待同步 1=已同步 2=已失效';
CREATE INDEX IF NOT EXISTS idx_ont_ch_child  ON ont_class_hierarchies (child_class_iri);
CREATE INDEX IF NOT EXISTS idx_ont_ch_parent ON ont_class_hierarchies (parent_class_iri);

-- ------------------------------------------------------------
-- 种子 1：rated 分组属性模板（供骨架引用，source='builtin'）
-- ------------------------------------------------------------
INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'ratedFlow', 'datatype', '额定流量', 'rated', 'decimal', 0, 'http://qudt.org/vocab/unit/M3-PER-SEC', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'ratedFlow' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'ratedHead', 'datatype', '额定扬程', 'rated', 'decimal', 0, 'http://qudt.org/vocab/unit/M', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'ratedHead' AND deleted_at IS NULL);

INSERT INTO ont_property_templates (template_code, kind, label, category, type, is_identifier, unit_ref, source)
SELECT 'ratedPower', 'datatype', '额定功率', 'rated', 'decimal', 0, 'http://qudt.org/vocab/unit/KiloW', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_property_templates WHERE template_code = 'ratedPower' AND deleted_at IS NULL);

-- ------------------------------------------------------------
-- 种子 2：equipment 分类编码规则
-- ------------------------------------------------------------
INSERT INTO ont_classification_rules (tree_root, separator, level_digits, base_number, zero_pad, description)
SELECT 'equipment', '-', 2, 30, 1, '设备分类树编码规则'
WHERE NOT EXISTS (SELECT 1 FROM ont_classification_rules WHERE tree_root = 'equipment' AND deleted_at IS NULL);

-- ------------------------------------------------------------
-- 种子 3：equipment 分类树（先父后子，parent_id/parent_path 按编码子查询取）
-- ------------------------------------------------------------
INSERT INTO ont_class_templates (template_code, classification_code, label, label_cn, tree_root, parent_id, parent_path, tree_level, sort, icon, color, source)
SELECT 'equipment', '30', '设备', '设备', 'equipment', 0, '0,', 0, 1, '🔧', '#1677ff', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_class_templates WHERE template_code = 'equipment' AND deleted_at IS NULL);

INSERT INTO ont_class_templates (template_code, classification_code, label, label_cn, tree_root, parent_id, parent_path, tree_level, sort, icon, color, source)
SELECT 'pump', '30-01', '泵', '泵', 'equipment',
       (SELECT id FROM ont_class_templates WHERE template_code = 'equipment' AND deleted_at IS NULL),
       '0,' || (SELECT id FROM ont_class_templates WHERE template_code = 'equipment' AND deleted_at IS NULL) || ',', 1, 1, '💧', '#1677ff', 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL);

INSERT INTO ont_class_templates (template_code, classification_code, label, label_cn, tree_root, parent_id, parent_path, tree_level, sort, source)
SELECT 'spray-pump', '30-01-01', '喷淋泵', '喷淋泵', 'equipment',
       (SELECT id FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL),
       (SELECT parent_path || id || ',' FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL), 2, 1, 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_class_templates WHERE template_code = 'spray-pump' AND deleted_at IS NULL);

INSERT INTO ont_class_templates (template_code, classification_code, label, label_cn, tree_root, parent_id, parent_path, tree_level, sort, source)
SELECT 'feedwater-pump', '30-01-02', '给水泵', '给水泵', 'equipment',
       (SELECT id FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL),
       (SELECT parent_path || id || ',' FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL), 2, 2, 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_class_templates WHERE template_code = 'feedwater-pump' AND deleted_at IS NULL);

INSERT INTO ont_class_templates (template_code, classification_code, label, label_cn, tree_root, parent_id, parent_path, tree_level, sort, source)
SELECT 'sewage-pump', '30-01-03', '污水泵', '污水泵', 'equipment',
       (SELECT id FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL),
       (SELECT parent_path || id || ',' FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL), 2, 3, 'builtin'
WHERE NOT EXISTS (SELECT 1 FROM ont_class_templates WHERE template_code = 'sewage-pump' AND deleted_at IS NULL);

-- ------------------------------------------------------------
-- 种子 4：结构骨架（pump 4 条；spray-pump 追加 isActive 构成继承视图 5 条用例）
-- ------------------------------------------------------------
INSERT INTO ont_class_template_refs (class_template_id, property_template_code, ref_type, sort_order)
SELECT (SELECT id FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL), x.code, 'property', x.sort
FROM (VALUES ('name', 1), ('ratedFlow', 2), ('ratedHead', 3), ('ratedPower', 4)) AS x(code, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM ont_class_template_refs r
    WHERE r.class_template_id = (SELECT id FROM ont_class_templates WHERE template_code = 'pump' AND deleted_at IS NULL)
      AND r.property_template_code = x.code AND r.deleted_at IS NULL);

INSERT INTO ont_class_template_refs (class_template_id, property_template_code, ref_type, sort_order)
SELECT (SELECT id FROM ont_class_templates WHERE template_code = 'spray-pump' AND deleted_at IS NULL), 'isActive', 'property', 1
WHERE NOT EXISTS (
    SELECT 1 FROM ont_class_template_refs r
    WHERE r.class_template_id = (SELECT id FROM ont_class_templates WHERE template_code = 'spray-pump' AND deleted_at IS NULL)
      AND r.property_template_code = 'isActive' AND r.deleted_at IS NULL);

-- ------------------------------------------------------------
-- 种子 5：字典 ont_ref_type
-- ------------------------------------------------------------
INSERT INTO sys_dictionaries (name, type, status, "desc", created_at, updated_at)
SELECT '本体引用类型', 'ont_ref_type', true, '骨架引用类型 property/relationship', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM sys_dictionaries WHERE type = 'ont_ref_type');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT '属性', 'property', true, 1, (SELECT id FROM sys_dictionaries WHERE type = 'ont_ref_type'), now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_ref_type')
      AND d.value = 'property');

INSERT INTO sys_dictionary_details (label, value, status, sort, sys_dictionary_id, created_at, updated_at)
SELECT '关系', 'relationship', true, 2, (SELECT id FROM sys_dictionaries WHERE type = 'ont_ref_type'), now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_dictionary_details d
    WHERE d.sys_dictionary_id = (SELECT id FROM sys_dictionaries WHERE type = 'ont_ref_type')
      AND d.value = 'relationship');
