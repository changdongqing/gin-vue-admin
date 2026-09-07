-- ============================================================
-- 迁移: 000010_create_report_platform.down.sql
-- 回滚: 先清理演示数据集种子（按 set_code 精确删除），再按建表逆序 DROP 8 张表
-- ============================================================

DELETE FROM report_data_set_params WHERE set_code IN ('demo_sys_users', 'demo_user_role_stats');
DELETE FROM report_data_sets WHERE set_code IN ('demo_sys_users', 'demo_user_role_stats');

DROP TABLE IF EXISTS report_analysis_configs;
DROP TABLE IF EXISTS report_analysis_reports;
DROP TABLE IF EXISTS report_excel_templates;
DROP TABLE IF EXISTS report_excel_reports;
DROP TABLE IF EXISTS report_data_set_transforms;
DROP TABLE IF EXISTS report_data_set_params;
DROP TABLE IF EXISTS report_data_sets;
DROP TABLE IF EXISTS report_data_sources;
