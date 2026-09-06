-- ============================================================
-- 迁移: 000009_create_ont_ext_binding.down.sql
-- 回滚: 按依赖逆序 DROP（对象 → 日志 → 绑定属性/子表 → 绑定 → 表 → 模块）
-- ============================================================

DROP TABLE IF EXISTS ont_object_relations;
DROP TABLE IF EXISTS ont_object_attr_values;
DROP TABLE IF EXISTS ont_objects;
DROP TABLE IF EXISTS ont_ext_sync_logs;
DROP TABLE IF EXISTS ont_ext_binding_properties;
DROP TABLE IF EXISTS ont_ext_binding_details;
DROP TABLE IF EXISTS ont_ext_bindings;
DROP TABLE IF EXISTS ont_ext_tables;
DROP TABLE IF EXISTS ont_ext_modules;
