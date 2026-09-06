-- ============================================================
-- 迁移: 000009_create_ont_ext_binding.up.sql
-- 目的: 本体建模·03 外部模块关联
--       注册表域: ont_ext_modules / ont_ext_tables
--       绑定域:   ont_ext_bindings / ont_ext_binding_details / ont_ext_binding_properties
--       同步日志: ont_ext_sync_logs（无软删列，日志不删）
--       对象落地: ont_objects / ont_object_attr_values / ont_object_relations
-- 注: 跨域引用无外键约束，由 Service 校验保证；状态枚举沿用前端本地 options，不种字典
-- ============================================================

-- ── 注册表域 ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ont_ext_modules (
    id           bigserial     PRIMARY KEY,
    module_code  varchar(64)   NOT NULL,
    name         varchar(128)  NOT NULL,
    service_name varchar(128)  NOT NULL DEFAULT '',
    description  varchar(512)  NOT NULL DEFAULT '',
    status       smallint      NOT NULL DEFAULT 0,
    sort_order   int           NOT NULL DEFAULT 0,
    created_by   varchar(64)   NOT NULL DEFAULT '',
    updated_by   varchar(64)   NOT NULL DEFAULT '',
    created_at   timestamptz   NOT NULL DEFAULT now(),
    updated_at   timestamptz   NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
COMMENT ON TABLE ont_ext_modules IS '外部模块注册（业务表逻辑分组）';
COMMENT ON COLUMN ont_ext_modules.status IS '0启用/1停用';
COMMENT ON COLUMN ont_ext_modules.created_by IS '创建人';
COMMENT ON COLUMN ont_ext_modules.updated_by IS '更新人';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_ext_module_code
    ON ont_ext_modules (module_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ont_ext_tables (
    id                  bigserial     PRIMARY KEY,
    module_id           bigint        NOT NULL,
    table_name          varchar(128)  NOT NULL,
    display_name        varchar(128)  NOT NULL DEFAULT '',
    pk_column           varchar(128)  NOT NULL DEFAULT 'id',
    pk_type             varchar(64)   NOT NULL DEFAULT '',
    deleted_column      varchar(128)  NOT NULL DEFAULT '',
    update_time_column  varchar(128)  NOT NULL DEFAULT '',
    supports_incremental smallint     NOT NULL DEFAULT 0,
    remark              varchar(512)  NOT NULL DEFAULT '',
    created_by          varchar(64)   NOT NULL DEFAULT '',
    updated_by          varchar(64)   NOT NULL DEFAULT '',
    created_at          timestamptz   NOT NULL DEFAULT now(),
    updated_at          timestamptz   NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);
COMMENT ON TABLE ont_ext_tables IS '外部表注册（物理表全局至多注册一次；deleted/update_time 列注册时探测回填）';
COMMENT ON COLUMN ont_ext_tables.deleted_column IS '业务表软删列（探测：优先deleted_at回退deleted；同步WHERE用）';
COMMENT ON COLUMN ont_ext_tables.update_time_column IS '水位列（探测：优先updated_at回退update_time；增量同步用）';
COMMENT ON COLUMN ont_ext_tables.supports_incremental IS '支持增量 0/1（有水位列即1）';
COMMENT ON COLUMN ont_ext_tables.created_by IS '创建人';
COMMENT ON COLUMN ont_ext_tables.updated_by IS '更新人';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_ext_table_name
    ON ont_ext_tables (table_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_ext_table_module
    ON ont_ext_tables (module_id) WHERE deleted_at IS NULL;

-- ── 绑定域 ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ont_ext_bindings (
    id                    bigserial     PRIMARY KEY,
    project_id            bigint        NOT NULL,
    class_id              bigint        NOT NULL,
    table_id              bigint        NOT NULL,
    key_column            varchar(128)  NOT NULL DEFAULT 'id',
    code_column           varchar(128)  NOT NULL DEFAULT '',
    name_column           varchar(128)  NOT NULL DEFAULT '',
    parent_column         varchar(128)  NOT NULL DEFAULT '',
    status_column         varchar(128)  NOT NULL DEFAULT '',
    status_active_value   varchar(64)   NOT NULL DEFAULT '',
    sync_mode             smallint      NOT NULL DEFAULT 1,
    sync_order            int           NOT NULL DEFAULT 100,
    include_disabled      smallint      NOT NULL DEFAULT 0,
    conflict_strategy     smallint      NOT NULL DEFAULT 2,
    missing_target_policy smallint      NOT NULL DEFAULT 1,
    binding_status        smallint      NOT NULL DEFAULT 0,
    telemetry_enabled     smallint      NOT NULL DEFAULT 0,
    watermark             timestamptz,
    last_sync_time        timestamptz,
    last_sync_summary     varchar(512)  NOT NULL DEFAULT '',
    remark                varchar(512)  NOT NULL DEFAULT '',
    created_by            varchar(64)   NOT NULL DEFAULT '',
    updated_by            varchar(64)   NOT NULL DEFAULT '',
    created_at            timestamptz   NOT NULL DEFAULT now(),
    updated_at            timestamptz   NOT NULL DEFAULT now(),
    deleted_at            timestamptz
);
COMMENT ON TABLE ont_ext_bindings IS '类绑定配置（类↔主表；一个类至多一条生效绑定）';
COMMENT ON COLUMN ont_ext_bindings.sync_mode IS '1手动/2定时+手动(调度预留)';
COMMENT ON COLUMN ont_ext_bindings.conflict_strategy IS '值漂移 1仅报告/2业务表覆盖';
COMMENT ON COLUMN ont_ext_bindings.missing_target_policy IS '目标缺失 1跳过记issue/2整行失败';
COMMENT ON COLUMN ont_ext_bindings.binding_status IS '0草稿/1生效/2停用';
COMMENT ON COLUMN ont_ext_bindings.telemetry_enabled IS '遥测直读（预留）';
COMMENT ON COLUMN ont_ext_bindings.created_by IS '创建人';
COMMENT ON COLUMN ont_ext_bindings.updated_by IS '更新人';

CREATE INDEX IF NOT EXISTS idx_ont_ext_binding_project
    ON ont_ext_bindings (project_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_ext_binding_class
    ON ont_ext_bindings (class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_ext_binding_status
    ON ont_ext_bindings (binding_status) WHERE deleted_at IS NULL;
-- 一个类至多一条生效绑定
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_ext_binding_class_active
    ON ont_ext_bindings (class_id) WHERE binding_status = 1 AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ont_ext_binding_details (
    id                 bigserial     PRIMARY KEY,
    binding_id         bigint        NOT NULL,
    table_id           bigint        NOT NULL,
    detail_kind        smallint      NOT NULL DEFAULT 1,
    join_column        varchar(128)  NOT NULL DEFAULT '',
    identifier_column  varchar(128)  NOT NULL DEFAULT '',
    value_column       varchar(128)  NOT NULL DEFAULT '',
    ts_column          varchar(128)  NOT NULL DEFAULT '',
    sort_order         int           NOT NULL DEFAULT 0,
    remark             varchar(512)  NOT NULL DEFAULT '',
    created_at         timestamptz   NOT NULL DEFAULT now(),
    updated_at         timestamptz   NOT NULL DEFAULT now(),
    deleted_at         timestamptz
);
COMMENT ON TABLE ont_ext_binding_details IS '子表绑定（宽表=一对一扩展行；窄表=EAV一行一属性）';
COMMENT ON COLUMN ont_ext_binding_details.detail_kind IS '1宽表一对一/2窄表EAV';
COMMENT ON COLUMN ont_ext_binding_details.join_column IS '回连外键列（子表列）';
COMMENT ON COLUMN ont_ext_binding_details.identifier_column IS '窄表：属性标识列';
COMMENT ON COLUMN ont_ext_binding_details.value_column IS '窄表：值列（宽表可空）';

CREATE INDEX IF NOT EXISTS idx_ont_ext_bd_binding
    ON ont_ext_binding_details (binding_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ont_ext_binding_properties (
    id                 bigserial     PRIMARY KEY,
    binding_id         bigint        NOT NULL,
    binding_type       smallint      NOT NULL DEFAULT 1,
    property_id        bigint        NOT NULL,
    detail_binding_id  bigint,
    biz_column         varchar(128)  NOT NULL DEFAULT '',
    value_converter    smallint      NOT NULL DEFAULT 0,
    const_value        varchar(512)  NOT NULL DEFAULT '',
    enabled            smallint      NOT NULL DEFAULT 1,
    sort_order         int           NOT NULL DEFAULT 0,
    remark             varchar(512)  NOT NULL DEFAULT '',
    created_at         timestamptz   NOT NULL DEFAULT now(),
    updated_at         timestamptz   NOT NULL DEFAULT now(),
    deleted_at         timestamptz
);
COMMENT ON TABLE ont_ext_binding_properties IS '属性绑定（类属性↔业务列；窄表来源 biz_column 存 identifier 匹配值）';
COMMENT ON COLUMN ont_ext_binding_properties.binding_type IS '1数据属性/2对象属性';
COMMENT ON COLUMN ont_ext_binding_properties.property_id IS '→ ont_model_datatype_properties.id（type=1）或 ont_model_object_properties.id（type=2）';
COMMENT ON COLUMN ont_ext_binding_properties.detail_binding_id IS '值来源子表（NULL=主表）';
COMMENT ON COLUMN ont_ext_binding_properties.biz_column IS '主/宽表=列名；窄表=identifier匹配值';

CREATE INDEX IF NOT EXISTS idx_ont_ext_bp_binding
    ON ont_ext_binding_properties (binding_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_ext_bp_property
    ON ont_ext_binding_properties (property_id) WHERE deleted_at IS NULL;

-- ── 同步日志（无软删列：日志不删；索引不带 deleted_at 子句） ──
CREATE TABLE IF NOT EXISTS ont_ext_sync_logs (
    id            bigserial     PRIMARY KEY,
    binding_id    bigint        NOT NULL,
    class_id      bigint        NOT NULL,
    trigger_type  smallint      NOT NULL DEFAULT 2,
    sync_scope    smallint      NOT NULL DEFAULT 1,
    total_rows    int           NOT NULL DEFAULT 0,
    created       int           NOT NULL DEFAULT 0,
    updated       int           NOT NULL DEFAULT 0,
    skipped       int           NOT NULL DEFAULT 0,
    failed        int           NOT NULL DEFAULT 0,
    issues_count  int           NOT NULL DEFAULT 0,
    status        smallint      NOT NULL DEFAULT 1,
    duration_ms   int           NOT NULL DEFAULT 0,
    summary       varchar(1000) NOT NULL DEFAULT '',
    skip_reasons  text          NOT NULL DEFAULT '',
    start_time    timestamptz   NOT NULL DEFAULT now(),
    end_time      timestamptz
);
COMMENT ON TABLE ont_ext_sync_logs IS '外部同步日志（一次触发一条；试运行不记录）';
COMMENT ON COLUMN ont_ext_sync_logs.trigger_type IS '1反向(预留)/2手动/3定时(预留)/4试运行(预留)';
COMMENT ON COLUMN ont_ext_sync_logs.sync_scope IS '1全量/2增量';
COMMENT ON COLUMN ont_ext_sync_logs.status IS '1成功/2部分失败/3失败';

CREATE INDEX IF NOT EXISTS idx_ont_ext_log_binding
    ON ont_ext_sync_logs (binding_id);
CREATE INDEX IF NOT EXISTS idx_ont_ext_log_class
    ON ont_ext_sync_logs (class_id);

-- ── 对象落地（同步写目标；完整对象管理为后续需求） ──────────
CREATE TABLE IF NOT EXISTS ont_objects (
    id             bigserial     PRIMARY KEY,
    project_id     bigint        NOT NULL,
    class_id       bigint        NOT NULL,
    object_code    varchar(128)  NOT NULL,
    object_name    varchar(255)  NOT NULL DEFAULT '',
    parent_id      bigint,
    state          smallint      NOT NULL DEFAULT 1,
    source_type    smallint      NOT NULL DEFAULT 2,
    biz_table      varchar(128)  NOT NULL DEFAULT '',
    biz_key        varchar(128)  NOT NULL DEFAULT '',
    last_sync_time timestamptz,
    created_by     varchar(64)   NOT NULL DEFAULT '',
    updated_by     varchar(64)   NOT NULL DEFAULT '',
    created_at     timestamptz   NOT NULL DEFAULT now(),
    updated_at     timestamptz   NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);
COMMENT ON TABLE ont_objects IS '本体对象个体（ABox；本期由同步物化，对象管理页后续交付）';
COMMENT ON COLUMN ont_objects.state IS '1在用/2停用(状态列非在用值)/3失联(全量孤儿)';
COMMENT ON COLUMN ont_objects.source_type IS '1本体系(预留)/2业务同步';
COMMENT ON COLUMN ont_objects.created_by IS '创建人';
COMMENT ON COLUMN ont_objects.updated_by IS '更新人';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_objects_code
    ON ont_objects (project_id, object_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_objects_biz
    ON ont_objects (class_id, biz_table, biz_key)
    WHERE deleted_at IS NULL AND biz_table <> '';

CREATE TABLE IF NOT EXISTS ont_object_attr_values (
    id           bigserial     PRIMARY KEY,
    object_id    bigint        NOT NULL,
    property_id  bigint        NOT NULL,
    value_json   text          NOT NULL DEFAULT '',
    created_at   timestamptz   NOT NULL DEFAULT now(),
    updated_at   timestamptz   NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
COMMENT ON TABLE ont_object_attr_values IS '对象数据属性值（EAV；对象属性值走 ont_object_relations）';
COMMENT ON COLUMN ont_object_attr_values.value_json IS '按 xsd 兼容组转换后的 JSON 标量';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_oav
    ON ont_object_attr_values (object_id, property_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ont_object_relations (
    id                  bigserial   PRIMARY KEY,
    object_id           bigint      NOT NULL,
    object_property_id  bigint      NOT NULL,
    target_object_id    bigint      NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);
COMMENT ON TABLE ont_object_relations IS '对象关系（对象属性实例化；来源业务表外键列解析）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_or
    ON ont_object_relations (object_id, object_property_id, target_object_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_or_target
    ON ont_object_relations (target_object_id) WHERE deleted_at IS NULL;
