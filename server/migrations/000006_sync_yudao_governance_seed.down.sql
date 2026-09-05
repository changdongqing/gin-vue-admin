-- ============================================================
-- 迁移: 000006_sync_yudao_governance_seed.down.sql
-- 回滚: 撤销芋道种子同步——硬删除本次插入的芋道最终态行（当前存活行），
--       并恢复被 up 软删除的既有行（ gin-vue-admin 适配版旧种子）。
-- 说明: down 会将全部软删行复位，仅适用于回滚本迁移的开发场景
-- ============================================================

DELETE FROM ont_property_templates WHERE deleted_at IS NULL;
DELETE FROM ont_class_templates WHERE deleted_at IS NULL;
DELETE FROM ont_class_template_refs WHERE deleted_at IS NULL;
DELETE FROM ont_classification_rules WHERE deleted_at IS NULL;
DELETE FROM ont_quantity_kinds WHERE deleted_at IS NULL;
DELETE FROM ont_units WHERE deleted_at IS NULL;
DELETE FROM ont_annotation_properties WHERE deleted_at IS NULL;

UPDATE ont_property_templates SET deleted_at = NULL;
UPDATE ont_class_templates SET deleted_at = NULL;
UPDATE ont_class_template_refs SET deleted_at = NULL;
UPDATE ont_classification_rules SET deleted_at = NULL;
UPDATE ont_quantity_kinds SET deleted_at = NULL;
UPDATE ont_units SET deleted_at = NULL;
UPDATE ont_annotation_properties SET deleted_at = NULL;
