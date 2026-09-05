-- ============================================================
-- 迁移: 000005_create_ont_annotation_property.down.sql
-- 回滚: 删除注释属性注册表；清理 ont_applies_to 字典种子
-- ============================================================

DROP TABLE IF EXISTS ont_annotation_properties;

DELETE FROM sys_dictionary_details
WHERE sys_dictionary_id IN (SELECT id FROM sys_dictionaries WHERE type = 'ont_applies_to');

DELETE FROM sys_dictionaries WHERE type = 'ont_applies_to';
