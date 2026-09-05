-- =====================================================================
-- 000001_baseline.up.sql
-- 基线迁移：由 pg_dump --schema-only 生成（gvadq 库 2026-09-05 快照）
-- 覆盖 gin-vue-admin 全部业务表（system/example/plugin 等 36 张表）。
-- 注意：
--   1. 表结构与 server/model/** 保持同步，由 gorm AutoMigrate 驱动演进；
--   2. 本基线之后的结构变更一律编写新的增量迁移（000002_xxx），
--      生产环境配合 disable-auto-migrate: true 使用；
--   3. 已存在的存量库不会执行本脚本，由程序自动将版本固化到 1。
-- =====================================================================

--
-- PostgreSQL database dump
--


-- Dumped from database version 18.4 (Debian 18.4-1.pgdg13+1)
-- Dumped by pg_dump version 18.4 (Debian 18.4-1.pgdg13+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: casbin_rule; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.casbin_rule (
    id bigint NOT NULL,
    ptype character varying(100),
    v0 character varying(100),
    v1 character varying(100),
    v2 character varying(100),
    v3 character varying(100),
    v4 character varying(100),
    v5 character varying(100)
);


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.casbin_rule_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.casbin_rule_id_seq OWNED BY public.casbin_rule.id;


--
-- Name: exa_attachment_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_attachment_category (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(255) DEFAULT NULL::character varying,
    pid bigint DEFAULT 0
);


--
-- Name: COLUMN exa_attachment_category.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_attachment_category.name IS '分类名称';


--
-- Name: COLUMN exa_attachment_category.pid; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_attachment_category.pid IS '父节点ID';


--
-- Name: exa_attachment_category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_attachment_category_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_attachment_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_attachment_category_id_seq OWNED BY public.exa_attachment_category.id;


--
-- Name: exa_customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_customers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    customer_name text,
    customer_phone_data text,
    sys_user_id bigint,
    sys_user_authority_id bigint
);


--
-- Name: COLUMN exa_customers.customer_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.customer_name IS '客户名';


--
-- Name: COLUMN exa_customers.customer_phone_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.customer_phone_data IS '客户手机号';


--
-- Name: COLUMN exa_customers.sys_user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.sys_user_id IS '管理ID';


--
-- Name: COLUMN exa_customers.sys_user_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.sys_user_authority_id IS '管理角色ID';


--
-- Name: exa_customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_customers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_customers_id_seq OWNED BY public.exa_customers.id;


--
-- Name: exa_file_chunks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_file_chunks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    exa_file_id bigint,
    file_chunk_number bigint,
    file_chunk_path text
);


--
-- Name: exa_file_chunks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_file_chunks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_file_chunks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_file_chunks_id_seq OWNED BY public.exa_file_chunks.id;


--
-- Name: exa_file_upload_and_downloads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_file_upload_and_downloads (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    class_id bigint DEFAULT 0,
    url text,
    tag text,
    key text
);


--
-- Name: COLUMN exa_file_upload_and_downloads.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.name IS '文件名';


--
-- Name: COLUMN exa_file_upload_and_downloads.class_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.class_id IS '分类id';


--
-- Name: COLUMN exa_file_upload_and_downloads.url; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.url IS '文件地址';


--
-- Name: COLUMN exa_file_upload_and_downloads.tag; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.tag IS '文件标签';


--
-- Name: COLUMN exa_file_upload_and_downloads.key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.key IS '编号';


--
-- Name: exa_file_upload_and_downloads_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_file_upload_and_downloads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_file_upload_and_downloads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_file_upload_and_downloads_id_seq OWNED BY public.exa_file_upload_and_downloads.id;


--
-- Name: exa_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_files (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    file_name text,
    file_md5 text,
    file_path text,
    chunk_total bigint,
    is_finish boolean
);


--
-- Name: exa_files_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_files_id_seq OWNED BY public.exa_files.id;


--
-- Name: gva_announcements_info; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gva_announcements_info (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text,
    content text,
    user_id bigint,
    attachments jsonb
);


--
-- Name: COLUMN gva_announcements_info.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.title IS '公告标题';


--
-- Name: COLUMN gva_announcements_info.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.content IS '公告内容';


--
-- Name: COLUMN gva_announcements_info.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.user_id IS '发布者';


--
-- Name: COLUMN gva_announcements_info.attachments; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.attachments IS '相关附件';


--
-- Name: gva_announcements_info_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gva_announcements_info_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gva_announcements_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gva_announcements_info_id_seq OWNED BY public.gva_announcements_info.id;


--
-- Name: jwt_blacklists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jwt_blacklists (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    jwt text
);


--
-- Name: COLUMN jwt_blacklists.jwt; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.jwt_blacklists.jwt IS 'jwt';


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jwt_blacklists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jwt_blacklists_id_seq OWNED BY public.jwt_blacklists.id;


--
-- Name: sys_ai_workflow_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_ai_workflow_sessions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint,
    tab character varying(32),
    title character varying(255),
    summary text,
    conversation_id character varying(255),
    message_id character varying(255),
    current_node_id character varying(64),
    settings jsonb,
    form_data jsonb,
    result_data jsonb,
    messages jsonb
);


--
-- Name: COLUMN sys_ai_workflow_sessions.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.user_id IS '用户ID';


--
-- Name: COLUMN sys_ai_workflow_sessions.tab; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.tab IS '会话类型';


--
-- Name: COLUMN sys_ai_workflow_sessions.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.title IS '会话标题';


--
-- Name: COLUMN sys_ai_workflow_sessions.summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.summary IS '摘要';


--
-- Name: COLUMN sys_ai_workflow_sessions.conversation_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.conversation_id IS 'Dify会话ID';


--
-- Name: COLUMN sys_ai_workflow_sessions.message_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.message_id IS 'Dify消息ID';


--
-- Name: COLUMN sys_ai_workflow_sessions.current_node_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.current_node_id IS '当前选中节点ID';


--
-- Name: COLUMN sys_ai_workflow_sessions.settings; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.settings IS '页面设置';


--
-- Name: COLUMN sys_ai_workflow_sessions.form_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.form_data IS '表单数据';


--
-- Name: COLUMN sys_ai_workflow_sessions.result_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.result_data IS '当前展示结果';


--
-- Name: COLUMN sys_ai_workflow_sessions.messages; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ai_workflow_sessions.messages IS '会话消息';


--
-- Name: sys_ai_workflow_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_ai_workflow_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_ai_workflow_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_ai_workflow_sessions_id_seq OWNED BY public.sys_ai_workflow_sessions.id;


--
-- Name: sys_api_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_api_tokens (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint,
    authority_id bigint,
    token text,
    status boolean DEFAULT true,
    expires_at timestamp with time zone,
    remark text
);


--
-- Name: COLUMN sys_api_tokens.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_api_tokens.user_id IS '用户ID';


--
-- Name: COLUMN sys_api_tokens.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_api_tokens.authority_id IS '角色ID';


--
-- Name: COLUMN sys_api_tokens.token; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_api_tokens.token IS 'Token';


--
-- Name: COLUMN sys_api_tokens.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_api_tokens.status IS '状态';


--
-- Name: COLUMN sys_api_tokens.expires_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_api_tokens.expires_at IS '过期时间';


--
-- Name: COLUMN sys_api_tokens.remark; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_api_tokens.remark IS '备注';


--
-- Name: sys_api_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_api_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_api_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_api_tokens_id_seq OWNED BY public.sys_api_tokens.id;


--
-- Name: sys_apis; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_apis (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    path text,
    description text,
    api_group text,
    method text DEFAULT 'POST'::text
);


--
-- Name: COLUMN sys_apis.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.path IS 'api路径';


--
-- Name: COLUMN sys_apis.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.description IS 'api中文描述';


--
-- Name: COLUMN sys_apis.api_group; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.api_group IS 'api组';


--
-- Name: COLUMN sys_apis.method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.method IS '方法';


--
-- Name: sys_apis_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_apis_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_apis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_apis_id_seq OWNED BY public.sys_apis.id;


--
-- Name: sys_authorities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_authorities (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    authority_id bigint NOT NULL,
    authority_name text,
    parent_id bigint,
    default_router text DEFAULT 'dashboard'::text,
    data_scope bigint DEFAULT 1
);


--
-- Name: COLUMN sys_authorities.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.authority_id IS '角色ID';


--
-- Name: COLUMN sys_authorities.authority_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.authority_name IS '角色名';


--
-- Name: COLUMN sys_authorities.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.parent_id IS '父角色ID';


--
-- Name: COLUMN sys_authorities.default_router; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.default_router IS '默认菜单';


--
-- Name: COLUMN sys_authorities.data_scope; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.data_scope IS '数据范围 1全部 2自定义 3本公司 4本部门及以下 5本部门 6仅本人';


--
-- Name: sys_authorities_authority_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_authorities_authority_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_authorities_authority_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_authorities_authority_id_seq OWNED BY public.sys_authorities.authority_id;


--
-- Name: sys_authority_btns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_authority_btns (
    authority_id bigint,
    sys_menu_id bigint,
    sys_base_menu_btn_id bigint
);


--
-- Name: COLUMN sys_authority_btns.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_btns.authority_id IS '角色ID';


--
-- Name: COLUMN sys_authority_btns.sys_menu_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_btns.sys_menu_id IS '菜单ID';


--
-- Name: COLUMN sys_authority_btns.sys_base_menu_btn_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_btns.sys_base_menu_btn_id IS '菜单按钮ID';


--
-- Name: sys_authority_data_scopes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_authority_data_scopes (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    authority_id bigint NOT NULL,
    scope_type character varying(16) NOT NULL,
    target_id bigint NOT NULL
);


--
-- Name: COLUMN sys_authority_data_scopes.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_data_scopes.authority_id IS '角色ID';


--
-- Name: COLUMN sys_authority_data_scopes.scope_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_data_scopes.scope_type IS '范围类型 company|department';


--
-- Name: COLUMN sys_authority_data_scopes.target_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_data_scopes.target_id IS '公司或部门ID';


--
-- Name: sys_authority_data_scopes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_authority_data_scopes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_authority_data_scopes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_authority_data_scopes_id_seq OWNED BY public.sys_authority_data_scopes.id;


--
-- Name: sys_authority_menus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_authority_menus (
    sys_base_menu_id bigint NOT NULL,
    sys_authority_authority_id bigint NOT NULL
);


--
-- Name: COLUMN sys_authority_menus.sys_authority_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_menus.sys_authority_authority_id IS '角色ID';


--
-- Name: sys_auto_code_histories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_auto_code_histories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    table_name text,
    package text,
    request text,
    struct_name text,
    abbreviation text,
    business_db text,
    description text,
    templates text,
    injections text,
    flag bigint,
    api_ids text,
    menu_id bigint,
    export_template_id bigint,
    package_id bigint
);


--
-- Name: COLUMN sys_auto_code_histories.table_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.table_name IS '表名';


--
-- Name: COLUMN sys_auto_code_histories.package; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.package IS '模块名或插件名';


--
-- Name: COLUMN sys_auto_code_histories.request; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.request IS '前端传入的结构化信息';


--
-- Name: COLUMN sys_auto_code_histories.struct_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.struct_name IS '结构体名称';


--
-- Name: COLUMN sys_auto_code_histories.abbreviation; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.abbreviation IS '结构体简称';


--
-- Name: COLUMN sys_auto_code_histories.business_db; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.business_db IS '业务库';


--
-- Name: COLUMN sys_auto_code_histories.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.description IS '结构体中文名';


--
-- Name: COLUMN sys_auto_code_histories.templates; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.templates IS '模板信息';


--
-- Name: COLUMN sys_auto_code_histories.injections; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.injections IS '注入信息';


--
-- Name: COLUMN sys_auto_code_histories.flag; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.flag IS '[0:创建,1:回滚]';


--
-- Name: COLUMN sys_auto_code_histories.api_ids; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.api_ids IS '关联API ID';


--
-- Name: COLUMN sys_auto_code_histories.menu_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.menu_id IS '菜单ID';


--
-- Name: COLUMN sys_auto_code_histories.export_template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.export_template_id IS '导出模板ID';


--
-- Name: COLUMN sys_auto_code_histories.package_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.package_id IS '包ID';


--
-- Name: sys_auto_code_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_auto_code_histories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_auto_code_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_auto_code_histories_id_seq OWNED BY public.sys_auto_code_histories.id;


--
-- Name: sys_auto_code_packages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_auto_code_packages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    "desc" text,
    label text,
    template text,
    package_name text,
    module text
);


--
-- Name: COLUMN sys_auto_code_packages."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages."desc" IS '描述';


--
-- Name: COLUMN sys_auto_code_packages.label; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages.label IS '显示名称';


--
-- Name: COLUMN sys_auto_code_packages.template; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages.template IS '模板';


--
-- Name: COLUMN sys_auto_code_packages.package_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages.package_name IS '包名';


--
-- Name: sys_auto_code_packages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_auto_code_packages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_auto_code_packages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_auto_code_packages_id_seq OWNED BY public.sys_auto_code_packages.id;


--
-- Name: sys_base_menu_btns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_base_menu_btns (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    "desc" text,
    sys_base_menu_id bigint
);


--
-- Name: COLUMN sys_base_menu_btns.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_btns.name IS '按钮关键key';


--
-- Name: COLUMN sys_base_menu_btns.sys_base_menu_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_btns.sys_base_menu_id IS '菜单ID';


--
-- Name: sys_base_menu_btns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_base_menu_btns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menu_btns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_base_menu_btns_id_seq OWNED BY public.sys_base_menu_btns.id;


--
-- Name: sys_base_menu_parameters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_base_menu_parameters (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    sys_base_menu_id bigint,
    type text,
    key text,
    value text
);


--
-- Name: COLUMN sys_base_menu_parameters.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_parameters.type IS '地址栏携带参数为params还是query';


--
-- Name: COLUMN sys_base_menu_parameters.key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_parameters.key IS '地址栏携带参数的key';


--
-- Name: COLUMN sys_base_menu_parameters.value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_parameters.value IS '地址栏携带参数的值';


--
-- Name: sys_base_menu_parameters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_base_menu_parameters_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menu_parameters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_base_menu_parameters_id_seq OWNED BY public.sys_base_menu_parameters.id;


--
-- Name: sys_base_menus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_base_menus (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    menu_level bigint,
    parent_id bigint,
    path text,
    name text,
    hidden boolean,
    component text,
    sort bigint,
    active_name text,
    keep_alive boolean,
    default_menu boolean,
    title text,
    icon text,
    close_tab boolean,
    transition_type text
);


--
-- Name: COLUMN sys_base_menus.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.parent_id IS '父菜单ID';


--
-- Name: COLUMN sys_base_menus.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.path IS '路由path';


--
-- Name: COLUMN sys_base_menus.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.name IS '路由name';


--
-- Name: COLUMN sys_base_menus.hidden; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.hidden IS '是否在列表隐藏';


--
-- Name: COLUMN sys_base_menus.component; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.component IS '对应前端文件路径';


--
-- Name: COLUMN sys_base_menus.sort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.sort IS '排序标记';


--
-- Name: COLUMN sys_base_menus.active_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.active_name IS '高亮菜单';


--
-- Name: COLUMN sys_base_menus.keep_alive; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.keep_alive IS '是否缓存';


--
-- Name: COLUMN sys_base_menus.default_menu; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.default_menu IS '是否是基础路由（开发中）';


--
-- Name: COLUMN sys_base_menus.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.title IS '菜单名';


--
-- Name: COLUMN sys_base_menus.icon; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.icon IS '菜单图标';


--
-- Name: COLUMN sys_base_menus.close_tab; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.close_tab IS '自动关闭tab';


--
-- Name: COLUMN sys_base_menus.transition_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.transition_type IS '路由切换动画';


--
-- Name: sys_base_menus_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_base_menus_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menus_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_base_menus_id_seq OWNED BY public.sys_base_menus.id;


--
-- Name: sys_companies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_companies (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(100) NOT NULL,
    code character varying(64) NOT NULL,
    parent_id bigint DEFAULT 0,
    parent_ids character varying(500),
    tree_level bigint DEFAULT 0,
    sort bigint DEFAULT 0,
    leader character varying(64),
    phone character varying(32),
    email character varying(128),
    address character varying(255),
    remarks character varying(255),
    status bigint DEFAULT 1
);


--
-- Name: COLUMN sys_companies.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.name IS '公司名称';


--
-- Name: COLUMN sys_companies.code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.code IS '公司编码';


--
-- Name: COLUMN sys_companies.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.parent_id IS '父公司ID 0为根';


--
-- Name: COLUMN sys_companies.parent_ids; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.parent_ids IS '物化路径 0,1,3,';


--
-- Name: COLUMN sys_companies.tree_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.tree_level IS '层级 根为0';


--
-- Name: COLUMN sys_companies.sort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.sort IS '排序';


--
-- Name: COLUMN sys_companies.leader; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.leader IS '负责人';


--
-- Name: COLUMN sys_companies.phone; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.phone IS '联系电话';


--
-- Name: COLUMN sys_companies.email; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.email IS '邮箱';


--
-- Name: COLUMN sys_companies.address; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.address IS '地址';


--
-- Name: COLUMN sys_companies.remarks; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.remarks IS '备注';


--
-- Name: COLUMN sys_companies.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_companies.status IS '状态 1启用 2停用';


--
-- Name: sys_companies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_companies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_companies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_companies_id_seq OWNED BY public.sys_companies.id;


--
-- Name: sys_data_authority_id; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_data_authority_id (
    sys_authority_authority_id bigint NOT NULL,
    data_authority_id_authority_id bigint NOT NULL
);


--
-- Name: COLUMN sys_data_authority_id.sys_authority_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_data_authority_id.sys_authority_authority_id IS '角色ID';


--
-- Name: COLUMN sys_data_authority_id.data_authority_id_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_data_authority_id.data_authority_id_authority_id IS '角色ID';


--
-- Name: sys_departments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_departments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    company_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    code character varying(64) NOT NULL,
    parent_id bigint DEFAULT 0,
    parent_ids character varying(500),
    tree_level bigint DEFAULT 0,
    sort bigint DEFAULT 0,
    leader character varying(64),
    phone character varying(32),
    email character varying(128),
    address character varying(255),
    remarks character varying(255),
    status bigint DEFAULT 1
);


--
-- Name: COLUMN sys_departments.company_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.company_id IS '挂靠公司ID';


--
-- Name: COLUMN sys_departments.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.name IS '部门名称';


--
-- Name: COLUMN sys_departments.code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.code IS '部门编码';


--
-- Name: COLUMN sys_departments.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.parent_id IS '父部门ID 0为根';


--
-- Name: COLUMN sys_departments.parent_ids; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.parent_ids IS '物化路径 0,1,3,';


--
-- Name: COLUMN sys_departments.tree_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.tree_level IS '层级 根为0';


--
-- Name: COLUMN sys_departments.sort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.sort IS '排序';


--
-- Name: COLUMN sys_departments.leader; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.leader IS '负责人';


--
-- Name: COLUMN sys_departments.phone; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.phone IS '联系电话';


--
-- Name: COLUMN sys_departments.email; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.email IS '邮箱';


--
-- Name: COLUMN sys_departments.address; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.address IS '地址';


--
-- Name: COLUMN sys_departments.remarks; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.remarks IS '备注';


--
-- Name: COLUMN sys_departments.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_departments.status IS '状态 1启用 2停用';


--
-- Name: sys_departments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_departments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_departments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_departments_id_seq OWNED BY public.sys_departments.id;


--
-- Name: sys_dictionaries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_dictionaries (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    type text,
    status boolean,
    "desc" text,
    parent_id bigint
);


--
-- Name: COLUMN sys_dictionaries.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries.name IS '字典名（中）';


--
-- Name: COLUMN sys_dictionaries.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries.type IS '字典名（英）';


--
-- Name: COLUMN sys_dictionaries.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries.status IS '状态';


--
-- Name: COLUMN sys_dictionaries."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries."desc" IS '描述';


--
-- Name: COLUMN sys_dictionaries.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries.parent_id IS '父级字典ID';


--
-- Name: sys_dictionaries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_dictionaries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_dictionaries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_dictionaries_id_seq OWNED BY public.sys_dictionaries.id;


--
-- Name: sys_dictionary_details; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_dictionary_details (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    label text,
    value text,
    extend text,
    status boolean,
    sort bigint,
    sys_dictionary_id bigint,
    parent_id bigint,
    level bigint,
    path text
);


--
-- Name: COLUMN sys_dictionary_details.label; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.label IS '展示值';


--
-- Name: COLUMN sys_dictionary_details.value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.value IS '字典值';


--
-- Name: COLUMN sys_dictionary_details.extend; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.extend IS '扩展值';


--
-- Name: COLUMN sys_dictionary_details.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.status IS '启用状态';


--
-- Name: COLUMN sys_dictionary_details.sort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.sort IS '排序标记';


--
-- Name: COLUMN sys_dictionary_details.sys_dictionary_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.sys_dictionary_id IS '关联标记';


--
-- Name: COLUMN sys_dictionary_details.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.parent_id IS '父级字典详情ID';


--
-- Name: COLUMN sys_dictionary_details.level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.level IS '层级深度';


--
-- Name: COLUMN sys_dictionary_details.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.path IS '层级路径';


--
-- Name: sys_dictionary_details_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_dictionary_details_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_dictionary_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_dictionary_details_id_seq OWNED BY public.sys_dictionary_details.id;


--
-- Name: sys_error; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_error (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    form text,
    info text,
    level text,
    solution text,
    status character varying(20) DEFAULT '未处理'::character varying
);


--
-- Name: COLUMN sys_error.form; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_error.form IS '错误来源';


--
-- Name: COLUMN sys_error.info; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_error.info IS '错误内容';


--
-- Name: COLUMN sys_error.level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_error.level IS '日志等级';


--
-- Name: COLUMN sys_error.solution; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_error.solution IS '解决方案';


--
-- Name: COLUMN sys_error.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_error.status IS '处理状态';


--
-- Name: sys_error_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_error_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_error_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_error_id_seq OWNED BY public.sys_error.id;


--
-- Name: sys_export_template_condition; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_export_template_condition (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    template_id text,
    "from" text,
    "column" text,
    operator text
);


--
-- Name: COLUMN sys_export_template_condition.template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition.template_id IS '模板标识';


--
-- Name: COLUMN sys_export_template_condition."from"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition."from" IS '条件取的key';


--
-- Name: COLUMN sys_export_template_condition."column"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition."column" IS '作为查询条件的字段';


--
-- Name: COLUMN sys_export_template_condition.operator; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition.operator IS '操作符';


--
-- Name: sys_export_template_condition_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_export_template_condition_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_template_condition_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_export_template_condition_id_seq OWNED BY public.sys_export_template_condition.id;


--
-- Name: sys_export_template_join; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_export_template_join (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    template_id text,
    joins text,
    "table" text,
    "on" text
);


--
-- Name: COLUMN sys_export_template_join.template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join.template_id IS '模板标识';


--
-- Name: COLUMN sys_export_template_join.joins; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join.joins IS '关联';


--
-- Name: COLUMN sys_export_template_join."table"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join."table" IS '关联表';


--
-- Name: COLUMN sys_export_template_join."on"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join."on" IS '关联条件';


--
-- Name: sys_export_template_join_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_export_template_join_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_template_join_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_export_template_join_id_seq OWNED BY public.sys_export_template_join.id;


--
-- Name: sys_export_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_export_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    db_name text,
    name text,
    table_name text,
    template_id text,
    template_info text,
    sql text,
    import_sql text,
    "limit" bigint,
    "order" text
);


--
-- Name: COLUMN sys_export_templates.db_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.db_name IS '数据库名称';


--
-- Name: COLUMN sys_export_templates.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.name IS '模板名称';


--
-- Name: COLUMN sys_export_templates.table_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.table_name IS '表名称';


--
-- Name: COLUMN sys_export_templates.template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.template_id IS '模板标识';


--
-- Name: COLUMN sys_export_templates.sql; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.sql IS '自定义导出SQL';


--
-- Name: COLUMN sys_export_templates.import_sql; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.import_sql IS '自定义导入SQL';


--
-- Name: COLUMN sys_export_templates."limit"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates."limit" IS '导出限制';


--
-- Name: COLUMN sys_export_templates."order"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates."order" IS '排序';


--
-- Name: sys_export_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_export_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_export_templates_id_seq OWNED BY public.sys_export_templates.id;


--
-- Name: sys_ignore_apis; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_ignore_apis (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    path text,
    method text DEFAULT 'POST'::text
);


--
-- Name: COLUMN sys_ignore_apis.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ignore_apis.path IS 'api路径';


--
-- Name: COLUMN sys_ignore_apis.method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ignore_apis.method IS '方法';


--
-- Name: sys_ignore_apis_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_ignore_apis_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_ignore_apis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_ignore_apis_id_seq OWNED BY public.sys_ignore_apis.id;


--
-- Name: sys_login_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_login_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    username text,
    ip text,
    status boolean,
    error_message text,
    agent text,
    user_id bigint
);


--
-- Name: COLUMN sys_login_logs.username; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_login_logs.username IS '用户名';


--
-- Name: COLUMN sys_login_logs.ip; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_login_logs.ip IS '请求ip';


--
-- Name: COLUMN sys_login_logs.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_login_logs.status IS '登录状态';


--
-- Name: COLUMN sys_login_logs.error_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_login_logs.error_message IS '错误信息';


--
-- Name: COLUMN sys_login_logs.agent; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_login_logs.agent IS '代理';


--
-- Name: COLUMN sys_login_logs.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_login_logs.user_id IS '用户id';


--
-- Name: sys_login_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_login_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_login_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_login_logs_id_seq OWNED BY public.sys_login_logs.id;


--
-- Name: sys_operation_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_operation_records (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ip text,
    method text,
    path text,
    status bigint,
    latency bigint,
    agent text,
    error_message text,
    body text,
    resp text,
    user_id bigint
);


--
-- Name: COLUMN sys_operation_records.ip; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.ip IS '请求ip';


--
-- Name: COLUMN sys_operation_records.method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.method IS '请求方法';


--
-- Name: COLUMN sys_operation_records.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.path IS '请求路径';


--
-- Name: COLUMN sys_operation_records.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.status IS '请求状态';


--
-- Name: COLUMN sys_operation_records.latency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.latency IS '延迟';


--
-- Name: COLUMN sys_operation_records.agent; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.agent IS '代理';


--
-- Name: COLUMN sys_operation_records.error_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.error_message IS '错误信息';


--
-- Name: COLUMN sys_operation_records.body; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.body IS '请求Body';


--
-- Name: COLUMN sys_operation_records.resp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.resp IS '响应Body';


--
-- Name: COLUMN sys_operation_records.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.user_id IS '用户id';


--
-- Name: sys_operation_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_operation_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_operation_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_operation_records_id_seq OWNED BY public.sys_operation_records.id;


--
-- Name: sys_params; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_params (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    key text,
    value text,
    "desc" text
);


--
-- Name: COLUMN sys_params.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params.name IS '参数名称';


--
-- Name: COLUMN sys_params.key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params.key IS '参数键';


--
-- Name: COLUMN sys_params.value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params.value IS '参数值';


--
-- Name: COLUMN sys_params."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params."desc" IS '参数说明';


--
-- Name: sys_params_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_params_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_params_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_params_id_seq OWNED BY public.sys_params.id;


--
-- Name: sys_user_authority; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_user_authority (
    sys_user_id bigint NOT NULL,
    sys_authority_authority_id bigint NOT NULL
);


--
-- Name: COLUMN sys_user_authority.sys_authority_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_user_authority.sys_authority_authority_id IS '角色ID';


--
-- Name: sys_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    uuid text,
    username text,
    password text,
    nick_name text DEFAULT '系统用户'::text,
    header_img text DEFAULT 'https://qmplusimg.henrongyi.top/gva_header.jpg'::text,
    authority_id bigint DEFAULT 888,
    phone text,
    email text,
    enable bigint DEFAULT 1,
    origin_setting jsonb,
    department_id bigint DEFAULT 0
);


--
-- Name: COLUMN sys_users.uuid; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.uuid IS '用户UUID';


--
-- Name: COLUMN sys_users.username; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.username IS '用户登录名';


--
-- Name: COLUMN sys_users.password; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.password IS '用户登录密码';


--
-- Name: COLUMN sys_users.nick_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.nick_name IS '用户昵称';


--
-- Name: COLUMN sys_users.header_img; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.header_img IS '用户头像';


--
-- Name: COLUMN sys_users.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.authority_id IS '用户角色ID';


--
-- Name: COLUMN sys_users.phone; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.phone IS '用户手机号';


--
-- Name: COLUMN sys_users.email; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.email IS '用户邮箱';


--
-- Name: COLUMN sys_users.enable; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.enable IS '用户是否被冻结 1正常 2冻结';


--
-- Name: COLUMN sys_users.origin_setting; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.origin_setting IS '配置';


--
-- Name: COLUMN sys_users.department_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.department_id IS '主属部门ID 0为未分配';


--
-- Name: sys_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_users_id_seq OWNED BY public.sys_users.id;


--
-- Name: sys_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_versions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version_name character varying(255),
    version_code character varying(100),
    description character varying(500),
    version_data text
);


--
-- Name: COLUMN sys_versions.version_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_versions.version_name IS '版本名称';


--
-- Name: COLUMN sys_versions.version_code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_versions.version_code IS '版本号';


--
-- Name: COLUMN sys_versions.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_versions.description IS '版本描述';


--
-- Name: COLUMN sys_versions.version_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_versions.version_data IS '版本数据JSON';


--
-- Name: sys_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_versions_id_seq OWNED BY public.sys_versions.id;


--
-- Name: casbin_rule id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casbin_rule ALTER COLUMN id SET DEFAULT nextval('public.casbin_rule_id_seq'::regclass);


--
-- Name: exa_attachment_category id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_attachment_category ALTER COLUMN id SET DEFAULT nextval('public.exa_attachment_category_id_seq'::regclass);


--
-- Name: exa_customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_customers ALTER COLUMN id SET DEFAULT nextval('public.exa_customers_id_seq'::regclass);


--
-- Name: exa_file_chunks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_chunks ALTER COLUMN id SET DEFAULT nextval('public.exa_file_chunks_id_seq'::regclass);


--
-- Name: exa_file_upload_and_downloads id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_upload_and_downloads ALTER COLUMN id SET DEFAULT nextval('public.exa_file_upload_and_downloads_id_seq'::regclass);


--
-- Name: exa_files id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_files ALTER COLUMN id SET DEFAULT nextval('public.exa_files_id_seq'::regclass);


--
-- Name: gva_announcements_info id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gva_announcements_info ALTER COLUMN id SET DEFAULT nextval('public.gva_announcements_info_id_seq'::regclass);


--
-- Name: jwt_blacklists id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jwt_blacklists ALTER COLUMN id SET DEFAULT nextval('public.jwt_blacklists_id_seq'::regclass);


--
-- Name: sys_ai_workflow_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_ai_workflow_sessions ALTER COLUMN id SET DEFAULT nextval('public.sys_ai_workflow_sessions_id_seq'::regclass);


--
-- Name: sys_api_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_api_tokens ALTER COLUMN id SET DEFAULT nextval('public.sys_api_tokens_id_seq'::regclass);


--
-- Name: sys_apis id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_apis ALTER COLUMN id SET DEFAULT nextval('public.sys_apis_id_seq'::regclass);


--
-- Name: sys_authorities authority_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authorities ALTER COLUMN authority_id SET DEFAULT nextval('public.sys_authorities_authority_id_seq'::regclass);


--
-- Name: sys_authority_data_scopes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authority_data_scopes ALTER COLUMN id SET DEFAULT nextval('public.sys_authority_data_scopes_id_seq'::regclass);


--
-- Name: sys_auto_code_histories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_histories ALTER COLUMN id SET DEFAULT nextval('public.sys_auto_code_histories_id_seq'::regclass);


--
-- Name: sys_auto_code_packages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_packages ALTER COLUMN id SET DEFAULT nextval('public.sys_auto_code_packages_id_seq'::regclass);


--
-- Name: sys_base_menu_btns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_btns ALTER COLUMN id SET DEFAULT nextval('public.sys_base_menu_btns_id_seq'::regclass);


--
-- Name: sys_base_menu_parameters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_parameters ALTER COLUMN id SET DEFAULT nextval('public.sys_base_menu_parameters_id_seq'::regclass);


--
-- Name: sys_base_menus id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menus ALTER COLUMN id SET DEFAULT nextval('public.sys_base_menus_id_seq'::regclass);


--
-- Name: sys_companies id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_companies ALTER COLUMN id SET DEFAULT nextval('public.sys_companies_id_seq'::regclass);


--
-- Name: sys_departments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_departments ALTER COLUMN id SET DEFAULT nextval('public.sys_departments_id_seq'::regclass);


--
-- Name: sys_dictionaries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionaries ALTER COLUMN id SET DEFAULT nextval('public.sys_dictionaries_id_seq'::regclass);


--
-- Name: sys_dictionary_details id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionary_details ALTER COLUMN id SET DEFAULT nextval('public.sys_dictionary_details_id_seq'::regclass);


--
-- Name: sys_error id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_error ALTER COLUMN id SET DEFAULT nextval('public.sys_error_id_seq'::regclass);


--
-- Name: sys_export_template_condition id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_condition ALTER COLUMN id SET DEFAULT nextval('public.sys_export_template_condition_id_seq'::regclass);


--
-- Name: sys_export_template_join id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_join ALTER COLUMN id SET DEFAULT nextval('public.sys_export_template_join_id_seq'::regclass);


--
-- Name: sys_export_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_templates ALTER COLUMN id SET DEFAULT nextval('public.sys_export_templates_id_seq'::regclass);


--
-- Name: sys_ignore_apis id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_ignore_apis ALTER COLUMN id SET DEFAULT nextval('public.sys_ignore_apis_id_seq'::regclass);


--
-- Name: sys_login_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_login_logs ALTER COLUMN id SET DEFAULT nextval('public.sys_login_logs_id_seq'::regclass);


--
-- Name: sys_operation_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_operation_records ALTER COLUMN id SET DEFAULT nextval('public.sys_operation_records_id_seq'::regclass);


--
-- Name: sys_params id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_params ALTER COLUMN id SET DEFAULT nextval('public.sys_params_id_seq'::regclass);


--
-- Name: sys_users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_users ALTER COLUMN id SET DEFAULT nextval('public.sys_users_id_seq'::regclass);


--
-- Name: sys_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_versions ALTER COLUMN id SET DEFAULT nextval('public.sys_versions_id_seq'::regclass);


--
-- Name: casbin_rule casbin_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casbin_rule
    ADD CONSTRAINT casbin_rule_pkey PRIMARY KEY (id);


--
-- Name: exa_attachment_category exa_attachment_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_attachment_category
    ADD CONSTRAINT exa_attachment_category_pkey PRIMARY KEY (id);


--
-- Name: exa_customers exa_customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_customers
    ADD CONSTRAINT exa_customers_pkey PRIMARY KEY (id);


--
-- Name: exa_file_chunks exa_file_chunks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_chunks
    ADD CONSTRAINT exa_file_chunks_pkey PRIMARY KEY (id);


--
-- Name: exa_file_upload_and_downloads exa_file_upload_and_downloads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_upload_and_downloads
    ADD CONSTRAINT exa_file_upload_and_downloads_pkey PRIMARY KEY (id);


--
-- Name: exa_files exa_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_files
    ADD CONSTRAINT exa_files_pkey PRIMARY KEY (id);


--
-- Name: gva_announcements_info gva_announcements_info_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gva_announcements_info
    ADD CONSTRAINT gva_announcements_info_pkey PRIMARY KEY (id);


--
-- Name: jwt_blacklists jwt_blacklists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jwt_blacklists
    ADD CONSTRAINT jwt_blacklists_pkey PRIMARY KEY (id);


--
-- Name: sys_ai_workflow_sessions sys_ai_workflow_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_ai_workflow_sessions
    ADD CONSTRAINT sys_ai_workflow_sessions_pkey PRIMARY KEY (id);


--
-- Name: sys_api_tokens sys_api_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_api_tokens
    ADD CONSTRAINT sys_api_tokens_pkey PRIMARY KEY (id);


--
-- Name: sys_apis sys_apis_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_apis
    ADD CONSTRAINT sys_apis_pkey PRIMARY KEY (id);


--
-- Name: sys_authority_data_scopes sys_authority_data_scopes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authority_data_scopes
    ADD CONSTRAINT sys_authority_data_scopes_pkey PRIMARY KEY (id);


--
-- Name: sys_authority_menus sys_authority_menus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authority_menus
    ADD CONSTRAINT sys_authority_menus_pkey PRIMARY KEY (sys_base_menu_id, sys_authority_authority_id);


--
-- Name: sys_auto_code_histories sys_auto_code_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_histories
    ADD CONSTRAINT sys_auto_code_histories_pkey PRIMARY KEY (id);


--
-- Name: sys_auto_code_packages sys_auto_code_packages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_packages
    ADD CONSTRAINT sys_auto_code_packages_pkey PRIMARY KEY (id);


--
-- Name: sys_base_menu_btns sys_base_menu_btns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_btns
    ADD CONSTRAINT sys_base_menu_btns_pkey PRIMARY KEY (id);


--
-- Name: sys_base_menu_parameters sys_base_menu_parameters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_parameters
    ADD CONSTRAINT sys_base_menu_parameters_pkey PRIMARY KEY (id);


--
-- Name: sys_base_menus sys_base_menus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menus
    ADD CONSTRAINT sys_base_menus_pkey PRIMARY KEY (id);


--
-- Name: sys_companies sys_companies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_companies
    ADD CONSTRAINT sys_companies_pkey PRIMARY KEY (id);


--
-- Name: sys_data_authority_id sys_data_authority_id_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_data_authority_id
    ADD CONSTRAINT sys_data_authority_id_pkey PRIMARY KEY (sys_authority_authority_id, data_authority_id_authority_id);


--
-- Name: sys_departments sys_departments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_departments
    ADD CONSTRAINT sys_departments_pkey PRIMARY KEY (id);


--
-- Name: sys_dictionaries sys_dictionaries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionaries
    ADD CONSTRAINT sys_dictionaries_pkey PRIMARY KEY (id);


--
-- Name: sys_dictionary_details sys_dictionary_details_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionary_details
    ADD CONSTRAINT sys_dictionary_details_pkey PRIMARY KEY (id);


--
-- Name: sys_error sys_error_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_error
    ADD CONSTRAINT sys_error_pkey PRIMARY KEY (id);


--
-- Name: sys_export_template_condition sys_export_template_condition_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_condition
    ADD CONSTRAINT sys_export_template_condition_pkey PRIMARY KEY (id);


--
-- Name: sys_export_template_join sys_export_template_join_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_join
    ADD CONSTRAINT sys_export_template_join_pkey PRIMARY KEY (id);


--
-- Name: sys_export_templates sys_export_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_templates
    ADD CONSTRAINT sys_export_templates_pkey PRIMARY KEY (id);


--
-- Name: sys_ignore_apis sys_ignore_apis_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_ignore_apis
    ADD CONSTRAINT sys_ignore_apis_pkey PRIMARY KEY (id);


--
-- Name: sys_login_logs sys_login_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_login_logs
    ADD CONSTRAINT sys_login_logs_pkey PRIMARY KEY (id);


--
-- Name: sys_operation_records sys_operation_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_operation_records
    ADD CONSTRAINT sys_operation_records_pkey PRIMARY KEY (id);


--
-- Name: sys_params sys_params_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_params
    ADD CONSTRAINT sys_params_pkey PRIMARY KEY (id);


--
-- Name: sys_user_authority sys_user_authority_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_user_authority
    ADD CONSTRAINT sys_user_authority_pkey PRIMARY KEY (sys_user_id, sys_authority_authority_id);


--
-- Name: sys_users sys_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_users
    ADD CONSTRAINT sys_users_pkey PRIMARY KEY (id);


--
-- Name: sys_versions sys_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_versions
    ADD CONSTRAINT sys_versions_pkey PRIMARY KEY (id);


--
-- Name: sys_authorities uni_sys_authorities_authority_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authorities
    ADD CONSTRAINT uni_sys_authorities_authority_id PRIMARY KEY (authority_id);


--
-- Name: idx_casbin_rule; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_casbin_rule ON public.casbin_rule USING btree (ptype, v0, v1, v2, v3, v4, v5);


--
-- Name: idx_exa_attachment_category_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_attachment_category_deleted_at ON public.exa_attachment_category USING btree (deleted_at);


--
-- Name: idx_exa_customers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_customers_deleted_at ON public.exa_customers USING btree (deleted_at);


--
-- Name: idx_exa_file_chunks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_file_chunks_deleted_at ON public.exa_file_chunks USING btree (deleted_at);


--
-- Name: idx_exa_file_upload_and_downloads_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_file_upload_and_downloads_deleted_at ON public.exa_file_upload_and_downloads USING btree (deleted_at);


--
-- Name: idx_exa_files_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_files_deleted_at ON public.exa_files USING btree (deleted_at);


--
-- Name: idx_gva_announcements_info_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gva_announcements_info_deleted_at ON public.gva_announcements_info USING btree (deleted_at);


--
-- Name: idx_jwt_blacklists_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jwt_blacklists_deleted_at ON public.jwt_blacklists USING btree (deleted_at);


--
-- Name: idx_sys_ai_workflow_sessions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_ai_workflow_sessions_deleted_at ON public.sys_ai_workflow_sessions USING btree (deleted_at);


--
-- Name: idx_sys_ai_workflow_sessions_tab; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_ai_workflow_sessions_tab ON public.sys_ai_workflow_sessions USING btree (tab);


--
-- Name: idx_sys_ai_workflow_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_ai_workflow_sessions_user_id ON public.sys_ai_workflow_sessions USING btree (user_id);


--
-- Name: idx_sys_api_tokens_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_api_tokens_deleted_at ON public.sys_api_tokens USING btree (deleted_at);


--
-- Name: idx_sys_apis_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_apis_deleted_at ON public.sys_apis USING btree (deleted_at);


--
-- Name: idx_sys_authority_data_scopes_authority_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_authority_data_scopes_authority_id ON public.sys_authority_data_scopes USING btree (authority_id);


--
-- Name: idx_sys_authority_data_scopes_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_authority_data_scopes_deleted_at ON public.sys_authority_data_scopes USING btree (deleted_at);


--
-- Name: idx_sys_auto_code_histories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_auto_code_histories_deleted_at ON public.sys_auto_code_histories USING btree (deleted_at);


--
-- Name: idx_sys_auto_code_packages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_auto_code_packages_deleted_at ON public.sys_auto_code_packages USING btree (deleted_at);


--
-- Name: idx_sys_base_menu_btns_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_base_menu_btns_deleted_at ON public.sys_base_menu_btns USING btree (deleted_at);


--
-- Name: idx_sys_base_menu_parameters_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_base_menu_parameters_deleted_at ON public.sys_base_menu_parameters USING btree (deleted_at);


--
-- Name: idx_sys_base_menus_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_base_menus_deleted_at ON public.sys_base_menus USING btree (deleted_at);


--
-- Name: idx_sys_companies_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_sys_companies_code ON public.sys_companies USING btree (code);


--
-- Name: idx_sys_companies_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_companies_deleted_at ON public.sys_companies USING btree (deleted_at);


--
-- Name: idx_sys_companies_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_companies_parent_id ON public.sys_companies USING btree (parent_id);


--
-- Name: idx_sys_companies_parent_ids; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_companies_parent_ids ON public.sys_companies USING btree (parent_ids);


--
-- Name: idx_sys_departments_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_sys_departments_code ON public.sys_departments USING btree (code);


--
-- Name: idx_sys_departments_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_departments_company_id ON public.sys_departments USING btree (company_id);


--
-- Name: idx_sys_departments_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_departments_deleted_at ON public.sys_departments USING btree (deleted_at);


--
-- Name: idx_sys_departments_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_departments_parent_id ON public.sys_departments USING btree (parent_id);


--
-- Name: idx_sys_departments_parent_ids; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_departments_parent_ids ON public.sys_departments USING btree (parent_ids);


--
-- Name: idx_sys_dictionaries_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_dictionaries_deleted_at ON public.sys_dictionaries USING btree (deleted_at);


--
-- Name: idx_sys_dictionary_details_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_dictionary_details_deleted_at ON public.sys_dictionary_details USING btree (deleted_at);


--
-- Name: idx_sys_error_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_error_deleted_at ON public.sys_error USING btree (deleted_at);


--
-- Name: idx_sys_export_template_condition_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_export_template_condition_deleted_at ON public.sys_export_template_condition USING btree (deleted_at);


--
-- Name: idx_sys_export_template_join_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_export_template_join_deleted_at ON public.sys_export_template_join USING btree (deleted_at);


--
-- Name: idx_sys_export_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_export_templates_deleted_at ON public.sys_export_templates USING btree (deleted_at);


--
-- Name: idx_sys_ignore_apis_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_ignore_apis_deleted_at ON public.sys_ignore_apis USING btree (deleted_at);


--
-- Name: idx_sys_login_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_login_logs_deleted_at ON public.sys_login_logs USING btree (deleted_at);


--
-- Name: idx_sys_operation_records_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_operation_records_deleted_at ON public.sys_operation_records USING btree (deleted_at);


--
-- Name: idx_sys_params_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_params_deleted_at ON public.sys_params USING btree (deleted_at);


--
-- Name: idx_sys_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_users_deleted_at ON public.sys_users USING btree (deleted_at);


--
-- Name: idx_sys_users_department_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_users_department_id ON public.sys_users USING btree (department_id);


--
-- Name: idx_sys_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_users_username ON public.sys_users USING btree (username);


--
-- Name: idx_sys_users_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_users_uuid ON public.sys_users USING btree (uuid);


--
-- Name: idx_sys_versions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_versions_deleted_at ON public.sys_versions USING btree (deleted_at);


--
-- PostgreSQL database dump complete
--


