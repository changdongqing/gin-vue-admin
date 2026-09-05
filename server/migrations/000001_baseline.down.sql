-- =====================================================================
-- 000001_baseline.down.sql
-- 基线回滚：删除 gin-vue-admin 全部业务表（与 000001_baseline.up.sql 对应）
-- 注意：仅用于彻底回退/重建环境，会丢失全部数据！
-- =====================================================================

DROP TABLE IF EXISTS public.casbin_rule CASCADE;
DROP TABLE IF EXISTS public.exa_attachment_category CASCADE;
DROP TABLE IF EXISTS public.exa_customers CASCADE;
DROP TABLE IF EXISTS public.exa_file_chunks CASCADE;
DROP TABLE IF EXISTS public.exa_file_upload_and_downloads CASCADE;
DROP TABLE IF EXISTS public.exa_files CASCADE;
DROP TABLE IF EXISTS public.gva_announcements_info CASCADE;
DROP TABLE IF EXISTS public.jwt_blacklists CASCADE;
DROP TABLE IF EXISTS public.sys_ai_workflow_sessions CASCADE;
DROP TABLE IF EXISTS public.sys_api_tokens CASCADE;
DROP TABLE IF EXISTS public.sys_apis CASCADE;
DROP TABLE IF EXISTS public.sys_authorities CASCADE;
DROP TABLE IF EXISTS public.sys_authority_btns CASCADE;
DROP TABLE IF EXISTS public.sys_authority_data_scopes CASCADE;
DROP TABLE IF EXISTS public.sys_authority_menus CASCADE;
DROP TABLE IF EXISTS public.sys_auto_code_histories CASCADE;
DROP TABLE IF EXISTS public.sys_auto_code_packages CASCADE;
DROP TABLE IF EXISTS public.sys_base_menu_btns CASCADE;
DROP TABLE IF EXISTS public.sys_base_menu_parameters CASCADE;
DROP TABLE IF EXISTS public.sys_base_menus CASCADE;
DROP TABLE IF EXISTS public.sys_companies CASCADE;
DROP TABLE IF EXISTS public.sys_data_authority_id CASCADE;
DROP TABLE IF EXISTS public.sys_departments CASCADE;
DROP TABLE IF EXISTS public.sys_dictionaries CASCADE;
DROP TABLE IF EXISTS public.sys_dictionary_details CASCADE;
DROP TABLE IF EXISTS public.sys_error CASCADE;
DROP TABLE IF EXISTS public.sys_export_template_condition CASCADE;
DROP TABLE IF EXISTS public.sys_export_template_join CASCADE;
DROP TABLE IF EXISTS public.sys_export_templates CASCADE;
DROP TABLE IF EXISTS public.sys_ignore_apis CASCADE;
DROP TABLE IF EXISTS public.sys_login_logs CASCADE;
DROP TABLE IF EXISTS public.sys_operation_records CASCADE;
DROP TABLE IF EXISTS public.sys_params CASCADE;
DROP TABLE IF EXISTS public.sys_user_authority CASCADE;
DROP TABLE IF EXISTS public.sys_users CASCADE;
DROP TABLE IF EXISTS public.sys_versions CASCADE;
