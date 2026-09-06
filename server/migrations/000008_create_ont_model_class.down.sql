-- ============================================================
-- 迁移: 000008_create_ont_model_class.down.sql
-- 回滚: 按依赖逆序删除类建模四表；清理 ont_model_xsd_type 字典
-- ============================================================

DROP TABLE IF EXISTS ont_model_subclassofs;
DROP TABLE IF EXISTS ont_model_object_properties;
DROP TABLE IF EXISTS ont_model_datatype_properties;
DROP TABLE IF EXISTS ont_model_classes;

DELETE FROM sys_dictionary_details
WHERE sys_dictionary_id IN (SELECT id FROM sys_dictionaries WHERE type = 'ont_model_xsd_type');

DELETE FROM sys_dictionaries WHERE type = 'ont_model_xsd_type';
