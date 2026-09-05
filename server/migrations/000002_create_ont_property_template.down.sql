-- ============================================================
-- 迁移: 000002_create_ont_property_template.down.sql
-- 回滚: 删除属性模板表；清理字典种子（按 type 幂等清理）
-- ============================================================

DROP TABLE IF EXISTS ont_property_templates;

DELETE FROM sys_dictionary_details
WHERE sys_dictionary_id IN (SELECT id FROM sys_dictionaries WHERE type IN ('ont_property_kind', 'ont_source'));

DELETE FROM sys_dictionaries WHERE type IN ('ont_property_kind', 'ont_source');
