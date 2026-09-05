-- ============================================================
-- 迁移: 000004_create_ont_unit.down.sql
-- 回滚: DROP 单位表与量纲表（量纲只读，无其它清理项）
-- ============================================================

DROP TABLE IF EXISTS ont_units;
DROP TABLE IF EXISTS ont_quantity_kinds;
