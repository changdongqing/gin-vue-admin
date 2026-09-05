-- ============================================================
-- 迁移: 000003_create_ont_class_template.down.sql
-- 回滚: 按依赖逆序 DROP 四表；清理 ont_ref_type 字典与 000003 补种的 rated 属性模板
-- ============================================================

DROP TABLE IF EXISTS ont_class_hierarchies;
DROP TABLE IF EXISTS ont_classification_rules;
DROP TABLE IF EXISTS ont_class_template_refs;
DROP TABLE IF EXISTS ont_class_templates;

-- 字典种子回滚（幂等）
DELETE FROM sys_dictionary_details
WHERE sys_dictionary_id IN (SELECT id FROM sys_dictionaries WHERE type = 'ont_ref_type');
DELETE FROM sys_dictionaries WHERE type = 'ont_ref_type';

-- 本迁移补种的 rated 属性模板回滚（仅删 builtin，避免误伤治理创建的同名模板）
DELETE FROM ont_property_templates
WHERE source = 'builtin' AND template_code IN ('ratedFlow', 'ratedHead', 'ratedPower');
