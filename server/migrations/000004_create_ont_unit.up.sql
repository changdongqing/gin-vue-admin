-- ============================================================
-- 迁移: 000004_create_ont_unit.up.sql
-- 目的: 量纲（只读）+ 单位（含线性换算参数）两表 + 种子（8 量纲 + 25 单位）
-- 说明: 换算模型 基准值 = 数值 × multiplier + offset；
--       温度参数按设计文档 §3.5 勘误口径（摄氏度为隐式基准）：
--       DEG_C(1, 0) / DEG_F(0.555555555555555, -17.77777777777776) / K(1, -273.15)
--       （DEG_F offset 取 -(32×乘数) 以保证 32°F→°C 精确为 0，与 AC 自洽）
-- ============================================================

CREATE TABLE IF NOT EXISTS ont_quantity_kinds (
    id                 bigserial    PRIMARY KEY,
    quantity_kind_code varchar(64)  NOT NULL,
    qudt_iri           varchar(255) NOT NULL,
    label              varchar(64)  NOT NULL,
    label_cn           varchar(64)  NOT NULL DEFAULT '',
    dimension_vector   varchar(32)  NOT NULL DEFAULT '',
    sort               int          NOT NULL DEFAULT 0,
    created_at         timestamptz  NOT NULL DEFAULT now(),
    updated_at         timestamptz  NOT NULL DEFAULT now(),
    deleted_at         timestamptz
);
COMMENT ON TABLE ont_quantity_kinds IS '本体量纲表（引用 QUDT，只读）';
COMMENT ON COLUMN ont_quantity_kinds.quantity_kind_code IS '量纲标识（如 Length/Mass）';
COMMENT ON COLUMN ont_quantity_kinds.qudt_iri IS 'QUDT 量纲IRI（唯一）';
COMMENT ON COLUMN ont_quantity_kinds.dimension_vector IS '量纲向量（如 A0E0L1I0M0H0T0D0）';
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_qk_code ON ont_quantity_kinds (quantity_kind_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_qk_iri  ON ont_quantity_kinds (qudt_iri) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ont_units (
    id                    bigserial       PRIMARY KEY,
    unit_code             varchar(64)     NOT NULL,
    qudt_iri              varchar(255)    NOT NULL,
    symbol                varchar(32)     NOT NULL DEFAULT '',
    label                 varchar(64)     NOT NULL,
    label_cn              varchar(64)     NOT NULL DEFAULT '',
    quantity_kind_code    varchar(64)     NOT NULL,
    conversion_multiplier numeric(30,15)  NOT NULL DEFAULT 1.0,
    conversion_offset     numeric(30,15)  NOT NULL DEFAULT 0.0,
    scaling_of            varchar(255)    NOT NULL DEFAULT '',
    ucum_code             varchar(32)     NOT NULL DEFAULT '',
    source                varchar(16)     NOT NULL DEFAULT 'custom',
    source_ref            varchar(64)     NOT NULL DEFAULT '',
    qudt_version          varchar(32)     NOT NULL DEFAULT '',
    status                smallint        NOT NULL DEFAULT 0,
    created_by            varchar(64)     NOT NULL DEFAULT '',
    updated_by            varchar(64)     NOT NULL DEFAULT '',
    created_at            timestamptz     NOT NULL DEFAULT now(),
    updated_at            timestamptz     NOT NULL DEFAULT now(),
    deleted_at            timestamptz
);
COMMENT ON TABLE ont_units IS '本体单位表（引用 QUDT，含线性换算参数）';
COMMENT ON COLUMN ont_units.unit_code IS '单位标识（如 KiloGM/M，唯一）';
COMMENT ON COLUMN ont_units.qudt_iri IS 'QUDT 单位IRI（唯一，换算/供给定位键）';
COMMENT ON COLUMN ont_units.symbol IS '符号（如 kg，冗余存储，QUDT 漂移时降级）';
COMMENT ON COLUMN ont_units.quantity_kind_code IS '所属量纲code（软引用 ont_quantity_kinds）';
COMMENT ON COLUMN ont_units.conversion_multiplier IS '换算乘数（基准值 = 数值×multiplier+offset）';
COMMENT ON COLUMN ont_units.conversion_offset IS '换算偏移（温度等有偏移单位）';
COMMENT ON COLUMN ont_units.scaling_of IS '基准单位 qudt_iri（派生关系留痕）';
COMMENT ON COLUMN ont_units.ucum_code IS 'UCUM 编码';
COMMENT ON COLUMN ont_units.source IS '来源: builtin/custom';
COMMENT ON COLUMN ont_units.qudt_version IS 'QUDT 版本快照（如 3.1.5）';
COMMENT ON COLUMN ont_units.status IS '业务启停: 0=正常 1=停用';
COMMENT ON COLUMN ont_units.created_by IS '创建人（用户名，API 层填充）';
COMMENT ON COLUMN ont_units.updated_by IS '更新人（用户名，API 层填充）';
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_unit_code ON ont_units (unit_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ont_unit_iri  ON ont_units (qudt_iri) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ont_unit_qk ON ont_units (quantity_kind_code) WHERE deleted_at IS NULL;

-- ------------------------------------------------------------
-- 量纲种子（8，只读）
-- ------------------------------------------------------------
INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'Length', 'http://qudt.org/vocab/quantitykind/Length', 'Length', '长度', 'A0E0L1I0M0H0T0D0', 1
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'Length' AND deleted_at IS NULL);

INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'Mass', 'http://qudt.org/vocab/quantitykind/Mass', 'Mass', '质量', 'A0E0L0I0M1H0T0D0', 2
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'Mass' AND deleted_at IS NULL);

INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'Time', 'http://qudt.org/vocab/quantitykind/Time', 'Time', '时间', 'A0E0L0I0M0H0T1D0', 3
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'Time' AND deleted_at IS NULL);

INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'Temperature', 'http://qudt.org/vocab/quantitykind/Temperature', 'Temperature', '温度', 'A0E0L0I0M0H1T0D0', 4
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'Temperature' AND deleted_at IS NULL);

INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'Area', 'http://qudt.org/vocab/quantitykind/Area', 'Area', '面积', 'A0E0L2I0M0H0T0D0', 5
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'Area' AND deleted_at IS NULL);

INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'Volume', 'http://qudt.org/vocab/quantitykind/Volume', 'Volume', '体积', 'A0E0L3I0M0H0T0D0', 6
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'Volume' AND deleted_at IS NULL);

INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'Currency', 'http://qudt.org/vocab/quantitykind/Currency', 'Currency', '货币', 'A0E0L0I0M0H0T0D0', 7
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'Currency' AND deleted_at IS NULL);

INSERT INTO ont_quantity_kinds (quantity_kind_code, qudt_iri, label, label_cn, dimension_vector, sort)
SELECT 'DataQuantity', 'http://qudt.org/vocab/quantitykind/DataQuantity', 'Data Quantity', '数据量', 'A0E0L0I0M0H0T0D1', 8
WHERE NOT EXISTS (SELECT 1 FROM ont_quantity_kinds WHERE quantity_kind_code = 'DataQuantity' AND deleted_at IS NULL);

-- ------------------------------------------------------------
-- 单位种子（source='builtin'，qudt_version='3.1.5'；各量纲隐式基准：
-- 长度=m 质量=kg 时间=s 温度=摄氏度 面积=m² 体积=m³ 货币=USD 数据量=byte）
-- 列序: (code, sym, lbl, lblcn, mult, scl, ucum)
-- ------------------------------------------------------------
-- Length
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, ucum_code, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'Length', x.mult, 0, x.scl, x.ucum, 'builtin', '3.1.5'
FROM (VALUES
    ('M',      'm',  'Metre',      '米',   1.0::numeric(30,15), '',                              'm'),
    ('CentiM', 'cm', 'Centimetre', '厘米', 0.01::numeric(30,15), 'http://qudt.org/vocab/unit/M', 'cm'),
    ('MilliM', 'mm', 'Millimetre', '毫米', 0.001::numeric(30,15), 'http://qudt.org/vocab/unit/M', 'mm'),
    ('KiloM',  'km', 'Kilometre',  '千米', 1000::numeric(30,15), 'http://qudt.org/vocab/unit/M',  'km')
) AS x(code, sym, lbl, lblcn, mult, scl, ucum)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);

-- Mass
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, ucum_code, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'Mass', x.mult, 0, x.scl, x.ucum, 'builtin', '3.1.5'
FROM (VALUES
    ('KiloGM',  'kg', 'Kilogram',  '千克', 1.0::numeric(30,15), '',                                       'kg'),
    ('GM',      'g',  'Gram',      '克',   0.001::numeric(30,15), 'http://qudt.org/vocab/unit/KiloGM',    'g'),
    ('MilliGM', 'mg', 'Milligram', '毫克', 0.000001::numeric(30,15), 'http://qudt.org/vocab/unit/KiloGM', 'mg')
) AS x(code, sym, lbl, lblcn, mult, scl, ucum)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);

-- Time
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, ucum_code, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'Time', x.mult, 0, x.scl, x.ucum, 'builtin', '3.1.5'
FROM (VALUES
    ('SEC', 's',   'Second', '秒', 1.0::numeric(30,15), '',                                's'),
    ('MIN', 'min', 'Minute', '分', 60::numeric(30,15), 'http://qudt.org/vocab/unit/SEC',  'min'),
    ('HR',  'h',   'Hour',   '时', 3600::numeric(30,15), 'http://qudt.org/vocab/unit/SEC', 'h')
) AS x(code, sym, lbl, lblcn, mult, scl, ucum)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);

-- Temperature（勘误口径：摄氏度为基准；DEG_F offset=-(32×乘数)，自洽验算 32°F→0°C / 100°C→373.15K / 373.15K→212°F）
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, ucum_code, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'Temperature', x.mult, x.off, x.scl, x.ucum, 'builtin', '3.1.5'
FROM (VALUES
    ('DEG_C', '°C', 'Degree Celsius',    '摄氏度', 1.0::numeric(30,15),               0.0::numeric(30,15),                '',                                 'Cel'),
    ('DEG_F', '°F', 'Degree Fahrenheit', '华氏度', 0.555555555555555::numeric(30,15), (-17.77777777777776)::numeric(30,15), 'http://qudt.org/vocab/unit/DEG_C', '[degF]'),
    ('K',     'K',  'Kelvin',            '开尔文', 1.0::numeric(30,15),               (-273.15)::numeric(30,15),          'http://qudt.org/vocab/unit/DEG_C', 'K')
) AS x(code, sym, lbl, lblcn, mult, off, scl, ucum)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);

-- Area
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, ucum_code, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'Area', x.mult, 0, x.scl, x.ucum, 'builtin', '3.1.5'
FROM (VALUES
    ('M2',      'm2', 'Square Metre', '平方米',   1.0::numeric(30,15), '',                            'm2'),
    ('FT2',     'ft2','Square Foot',  '平方英尺', 0.09290304::numeric(30,15), 'http://qudt.org/vocab/unit/M2', '[ft_i]2'),
    ('HECTARE', 'ha', 'Hectare',      '公顷',     10000::numeric(30,15), 'http://qudt.org/vocab/unit/M2', 'har')
) AS x(code, sym, lbl, lblcn, mult, scl, ucum)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);

-- Volume
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, ucum_code, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'Volume', x.mult, 0, x.scl, x.ucum, 'builtin', '3.1.5'
FROM (VALUES
    ('M3',  'm3',  'Cubic Metre',     '立方米',   1.0::numeric(30,15), '',                              'm3'),
    ('L',   'L',   'Litre',           '升',       0.001::numeric(30,15), 'http://qudt.org/vocab/unit/M3', 'L'),
    ('CM3', 'cm3', 'Cubic Centimetre','立方厘米', 0.000001::numeric(30,15), 'http://qudt.org/vocab/unit/M3', 'cm3')
) AS x(code, sym, lbl, lblcn, mult, scl, ucum)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);

-- Currency（乘数为参考汇率快照，非实时）
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'Currency', x.mult, 0, x.scl, 'builtin', '3.1.5'
FROM (VALUES
    ('USD', '$', 'US Dollar', '美元',   1.0::numeric(30,15), ''),
    ('CNY', '¥', 'Yuan',      '人民币', 0.14::numeric(30,15), 'http://qudt.org/vocab/unit/USD'),
    ('EUR', '€', 'Euro',      '欧元',   1.08::numeric(30,15), 'http://qudt.org/vocab/unit/USD')
) AS x(code, sym, lbl, lblcn, mult, scl)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);

-- DataQuantity
INSERT INTO ont_units (unit_code, qudt_iri, symbol, label, label_cn, quantity_kind_code, conversion_multiplier, conversion_offset, scaling_of, source, qudt_version)
SELECT x.code, 'http://qudt.org/vocab/unit/' || x.code, x.sym, x.lbl, x.lblcn, 'DataQuantity', x.mult, 0, x.scl, 'builtin', '3.1.5'
FROM (VALUES
    ('BYTE',     'B',  'Byte',     '字节',   1.0::numeric(30,15), ''),
    ('KiloBYTE', 'kB', 'Kilobyte', '千字节', 1024::numeric(30,15), 'http://qudt.org/vocab/unit/BYTE'),
    ('MegaBYTE', 'MB', 'Megabyte', '兆字节', 1048576::numeric(30,15), 'http://qudt.org/vocab/unit/BYTE')
) AS x(code, sym, lbl, lblcn, mult, scl)
WHERE NOT EXISTS (SELECT 1 FROM ont_units WHERE unit_code = x.code AND deleted_at IS NULL);
