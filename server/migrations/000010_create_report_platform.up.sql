-- ============================================================
-- 迁移: 000010_create_report_platform.up.sql
-- 目的: 报表平台 8 张表一次预建（对齐参考实现 V26 一次建表做法）
--       01 数据源:  report_data_sources
--       02 数据集:  report_data_sets / report_data_set_params / report_data_set_transforms
--       03 Excel:   report_excel_reports / report_excel_templates
--       05 分析:    report_analysis_reports / report_analysis_configs
-- 种子: 演示数据集（sys_users 查询，含 ${param} 与 <if param> 条件语法，携带 case_result，
--       供 03 设计器左栏字段与 05 白名单消费；演示数据源 demo_pg 由 Go 种子按本库连接配置创建）
-- 注: 业务唯一索引用 PG 部分索引（WHERE deleted_at IS NULL）+ Service 查重双保险；
--     跨域引用无外键约束，由 Service 校验保证；枚举沿用前端本地 options，不种字典
-- ============================================================

-- ── 01 数据源表 ───────────────────────────────────────────
CREATE TABLE IF NOT EXISTS report_data_sources (
    id          bigserial     PRIMARY KEY,
    source_code varchar(50)   NOT NULL,
    source_name varchar(100)  NOT NULL DEFAULT '',
    source_type varchar(50)   NOT NULL DEFAULT '',
    source_desc varchar(255)  NOT NULL DEFAULT '',
    source_config varchar(2048) NOT NULL DEFAULT '',
    enable_flag bool          NOT NULL DEFAULT true,
    created_by  varchar(64)   NOT NULL DEFAULT '',
    updated_by  varchar(64)   NOT NULL DEFAULT '',
    created_at  timestamptz   NOT NULL DEFAULT now(),
    updated_at  timestamptz   NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
COMMENT ON TABLE report_data_sources IS '报表平台-数据源表';
COMMENT ON COLUMN report_data_sources.source_code IS '数据源编码（唯一）';
COMMENT ON COLUMN report_data_sources.source_name IS '数据源名称';
COMMENT ON COLUMN report_data_sources.source_type IS '数据源类型（postgresql/mysql/sqlserver/.../http）';
COMMENT ON COLUMN report_data_sources.source_desc IS '描述';
COMMENT ON COLUMN report_data_sources.source_config IS '连接配置JSON（密码字段AES-GCM加密存储）';
COMMENT ON COLUMN report_data_sources.enable_flag IS '是否启用';

CREATE UNIQUE INDEX IF NOT EXISTS uk_report_data_source_code
    ON report_data_sources (source_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_data_source_type
    ON report_data_sources (source_type) WHERE deleted_at IS NULL;

-- ── 02 数据集主表 ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS report_data_sets (
    id           bigserial     PRIMARY KEY,
    set_code     varchar(50)   NOT NULL,
    set_name     varchar(100)  NOT NULL DEFAULT '',
    set_desc     varchar(255)  NOT NULL DEFAULT '',
    source_code  varchar(50)   NOT NULL DEFAULT '',
    set_type     varchar(10)   NOT NULL DEFAULT 'sql',
    dyn_sentence varchar(4096) NOT NULL DEFAULT '',
    case_result  text,
    enable_flag  bool          NOT NULL DEFAULT true,
    created_by   varchar(64)   NOT NULL DEFAULT '',
    updated_by   varchar(64)   NOT NULL DEFAULT '',
    created_at   timestamptz   NOT NULL DEFAULT now(),
    updated_at   timestamptz   NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
COMMENT ON TABLE report_data_sets IS '报表平台-数据集表';
COMMENT ON COLUMN report_data_sets.source_code IS '关联数据源编码（http 类型为空）';
COMMENT ON COLUMN report_data_sets.dyn_sentence IS '查询语句（SQL 含 ${param}/<if param>；HTTP 为请求配置JSON）';
COMMENT ON COLUMN report_data_sets.case_result IS '结果案例JSON（数据样例行，设计器字段来源，由种子/DBA维护）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_report_data_set_code
    ON report_data_sets (set_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_data_set_source
    ON report_data_sets (source_code) WHERE deleted_at IS NULL;

-- ── 02 数据集参数子表（set_code 逻辑关联）────────────────
CREATE TABLE IF NOT EXISTS report_data_set_params (
    id               bigserial     PRIMARY KEY,
    set_code         varchar(50)   NOT NULL,
    param_name       varchar(50)   NOT NULL DEFAULT '',
    param_desc       varchar(100)  NOT NULL DEFAULT '',
    param_type       varchar(20)   NOT NULL DEFAULT 'string',
    sample_item      varchar(1080) NOT NULL DEFAULT '',
    default_value    varchar(1080) NOT NULL DEFAULT '',
    dict_type        varchar(100)  NOT NULL DEFAULT '',
    custom_options   varchar(2048) NOT NULL DEFAULT '',
    date_format      varchar(50)   NOT NULL DEFAULT '',
    required_flag    bool          NOT NULL DEFAULT false,
    validation_rules varchar(2048) NOT NULL DEFAULT '',
    order_num        int           NOT NULL DEFAULT 0,
    created_by       varchar(64)   NOT NULL DEFAULT '',
    updated_by       varchar(64)   NOT NULL DEFAULT '',
    created_at       timestamptz   NOT NULL DEFAULT now(),
    updated_at       timestamptz   NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);
COMMENT ON TABLE report_data_set_params IS '报表平台-数据集参数表（set_code 关联）';
COMMENT ON COLUMN report_data_set_params.sample_item IS '示例值（测试预览/参数表单预填）';
COMMENT ON COLUMN report_data_set_params.default_value IS '默认值（直接值或 today/thisMonth 等表达式）';
COMMENT ON COLUMN report_data_set_params.validation_rules IS 'JS校验规则（预留，本期不执行）';

CREATE INDEX IF NOT EXISTS idx_report_data_set_param_set
    ON report_data_set_params (set_code) WHERE deleted_at IS NULL;

-- ── 02 数据集转换子表（set_code 逻辑关联）────────────────
CREATE TABLE IF NOT EXISTS report_data_set_transforms (
    id               bigserial   PRIMARY KEY,
    set_code         varchar(50) NOT NULL,
    transform_type   varchar(50) NOT NULL DEFAULT '',
    transform_script text,
    order_num        int         NOT NULL DEFAULT 0,
    created_by       varchar(64) NOT NULL DEFAULT '',
    updated_by       varchar(64) NOT NULL DEFAULT '',
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);
COMMENT ON TABLE report_data_set_transforms IS '报表平台-数据集转换表（set_code 关联）';
COMMENT ON COLUMN report_data_set_transforms.transform_type IS 'js（goja沙箱脚本）/ dict（字段值映射）';

CREATE INDEX IF NOT EXISTS idx_report_data_set_transform_set
    ON report_data_set_transforms (set_code) WHERE deleted_at IS NULL;

-- ── 03 Excel 报表元数据表 ─────────────────────────────────
CREATE TABLE IF NOT EXISTS report_excel_reports (
    id           bigserial     PRIMARY KEY,
    report_code  varchar(100)  NOT NULL,
    report_name  varchar(100)  NOT NULL DEFAULT '',
    report_group varchar(100)  NOT NULL DEFAULT '',
    report_desc  varchar(255)  NOT NULL DEFAULT '',
    created_by   varchar(64)   NOT NULL DEFAULT '',
    updated_by   varchar(64)   NOT NULL DEFAULT '',
    created_at   timestamptz   NOT NULL DEFAULT now(),
    updated_at   timestamptz   NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
COMMENT ON TABLE report_excel_reports IS '报表平台-Excel报表元数据表';
COMMENT ON COLUMN report_excel_reports.report_code IS '报表编码（唯一，创建后不可改）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_report_excel_report_code
    ON report_excel_reports (report_code) WHERE deleted_at IS NULL;

-- ── 03 Excel 报表模板内容表（与元数据按 report_code 1:1）──
CREATE TABLE IF NOT EXISTS report_excel_templates (
    id          bigserial    PRIMARY KEY,
    report_code varchar(100) NOT NULL,
    set_codes   varchar(500) NOT NULL DEFAULT '',
    set_param   text,
    json_str    text,
    created_by  varchar(64) NOT NULL DEFAULT '',
    updated_by  varchar(64) NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
COMMENT ON TABLE report_excel_templates IS '报表平台-Excel报表模板内容表';
COMMENT ON COLUMN report_excel_templates.set_codes IS '关联数据集编码（bind-datasets 维护，| 分隔）';
COMMENT ON COLUMN report_excel_templates.set_param IS '参数默认值JSON（预留）';
COMMENT ON COLUMN report_excel_templates.json_str IS 'Univer工作簿快照JSON（save-template 维护）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_report_excel_template_code
    ON report_excel_templates (report_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_excel_template_set_codes
    ON report_excel_templates (set_codes) WHERE deleted_at IS NULL;

-- ── 05 分析报表元数据表 ───────────────────────────────────
CREATE TABLE IF NOT EXISTS report_analysis_reports (
    id           bigserial     PRIMARY KEY,
    report_code  varchar(100)  NOT NULL,
    report_name  varchar(100)  NOT NULL DEFAULT '',
    report_group varchar(100)  NOT NULL DEFAULT '',
    report_desc  varchar(255)  NOT NULL DEFAULT '',
    set_code     varchar(50)   NOT NULL DEFAULT '',
    status       smallint      NOT NULL DEFAULT 0,
    created_by   varchar(64)   NOT NULL DEFAULT '',
    updated_by   varchar(64)   NOT NULL DEFAULT '',
    created_at   timestamptz   NOT NULL DEFAULT now(),
    updated_at   timestamptz   NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
COMMENT ON TABLE report_analysis_reports IS '报表平台-分析报表元数据表';
COMMENT ON COLUMN report_analysis_reports.set_code IS '关联数据集编码（单选）';
COMMENT ON COLUMN report_analysis_reports.status IS '0启用/1禁用';

CREATE UNIQUE INDEX IF NOT EXISTS uk_report_analysis_code
    ON report_analysis_reports (report_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_analysis_set
    ON report_analysis_reports (set_code) WHERE deleted_at IS NULL;

-- ── 05 分析报表配置内容表（与元数据按 report_code 1:1）────
CREATE TABLE IF NOT EXISTS report_analysis_configs (
    id          bigserial    PRIMARY KEY,
    report_code varchar(100) NOT NULL,
    config_json text,
    set_param   text,
    created_by  varchar(64) NOT NULL DEFAULT '',
    updated_by  varchar(64) NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
COMMENT ON TABLE report_analysis_configs IS '报表平台-分析报表配置内容表';
COMMENT ON COLUMN report_analysis_configs.config_json IS 'S2配置JSON（fields/options/theme 三段）';
COMMENT ON COLUMN report_analysis_configs.set_param IS '参数默认值JSON（预留，与Excel模板对称）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_report_analysis_config_code
    ON report_analysis_configs (report_code) WHERE deleted_at IS NULL;

-- ── 演示数据集种子（幂等；演示数据源 demo_pg 由 Go 种子按本库连接配置创建）──
INSERT INTO report_data_sets (set_code, set_name, set_desc, source_code, set_type, dyn_sentence, case_result, enable_flag, created_by, updated_by)
SELECT 'demo_sys_users', '系统用户清单', '演示数据集：sys_users 模糊查询（${param} 参数化 + <if param> 条件语法）', 'demo_pg', 'sql',
'SELECT id, username, nick_name, phone, email, created_at FROM sys_users WHERE 1=1
<if param="username"> AND username LIKE CONCAT(''%'', ${username}, ''%'')</if>
<if param="enable"> AND enable = ${enable}</if>
ORDER BY id',
'[{"id":1,"username":"admin","nick_name":"超级管理员","phone":"13800000001","email":"admin@example.com","created_at":"2026-01-01 08:00:00"},{"id":2,"username":"demo","nick_name":"演示用户","phone":"13800000002","email":"demo@example.com","created_at":"2026-01-02 08:00:00"}]',
true, 'seed', 'seed'
WHERE NOT EXISTS (SELECT 1 FROM report_data_sets WHERE set_code = 'demo_sys_users' AND deleted_at IS NULL);

INSERT INTO report_data_set_params (set_code, param_name, param_desc, param_type, sample_item, default_value, required_flag, order_num, created_by, updated_by)
SELECT v.set_code, v.param_name, v.param_desc, v.param_type, v.sample_item, v.default_value, v.required_flag, v.order_num, 'seed', 'seed'
FROM (VALUES
    ('demo_sys_users', 'username', '用户名（模糊匹配）', 'string', 'admin', '', false, 1),
    ('demo_sys_users', 'enable', '是否启用（1启用 0禁用）', 'number', '1', '', false, 2)
) AS v(set_code, param_name, param_desc, param_type, sample_item, default_value, required_flag, order_num)
WHERE NOT EXISTS (
    SELECT 1 FROM report_data_set_params p
    WHERE p.set_code = v.set_code AND p.param_name = v.param_name AND p.deleted_at IS NULL
);

INSERT INTO report_data_sets (set_code, set_name, set_desc, source_code, set_type, dyn_sentence, case_result, enable_flag, created_by, updated_by)
SELECT 'demo_user_role_stats', '用户角色分布', '演示数据集：sys_users 按角色预聚合（GROUP BY，供分析报表 NONE 直显演示）', 'demo_pg', 'sql',
'SELECT authority_id, COUNT(*) AS user_count FROM sys_users WHERE 1=1
<if param="authorityId"> AND authority_id = ${authorityId}</if>
GROUP BY authority_id ORDER BY authority_id',
'[{"authority_id":888,"user_count":1},{"authority_id":9528,"user_count":2}]',
true, 'seed', 'seed'
WHERE NOT EXISTS (SELECT 1 FROM report_data_sets WHERE set_code = 'demo_user_role_stats' AND deleted_at IS NULL);

INSERT INTO report_data_set_params (set_code, param_name, param_desc, param_type, sample_item, default_value, required_flag, order_num, created_by, updated_by)
SELECT v.set_code, v.param_name, v.param_desc, v.param_type, v.sample_item, v.default_value, v.required_flag, v.order_num, 'seed', 'seed'
FROM (VALUES
    ('demo_user_role_stats', 'authorityId', '角色ID（不填查全部）', 'number', '888', '', false, 1)
) AS v(set_code, param_name, param_desc, param_type, sample_item, default_value, required_flag, order_num)
WHERE NOT EXISTS (
    SELECT 1 FROM report_data_set_params p
    WHERE p.set_code = v.set_code AND p.param_name = v.param_name AND p.deleted_at IS NULL
);
