-- ============================================================
-- 迁移: 000007_create_ont_model_project.down.sql
-- 回滚: 删除本体项目/IRI前缀两张表；清理三个 ont_model_* 字典种子
-- ============================================================

DROP TABLE IF EXISTS ont_model_prefixes;
DROP TABLE IF EXISTS ont_model_projects;

DELETE FROM sys_dictionary_details
WHERE sys_dictionary_id IN (
    SELECT id FROM sys_dictionaries
    WHERE type IN ('ont_model_project_status', 'ont_model_format', 'ont_model_strategy'));

DELETE FROM sys_dictionaries
WHERE type IN ('ont_model_project_status', 'ont_model_format', 'ont_model_strategy');
