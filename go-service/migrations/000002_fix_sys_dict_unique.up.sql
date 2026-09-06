-- 000002 · 修复 sys_dict 上错误的单列唯一约束（缺陷 S-1）
--
-- alembic 的 77ea0fe08343_init_table.py 为 sys_dict 生成了一个单列 UNIQUE(dict_type)，
-- 与模型定义的复合语义 (dict_type, code) 矛盾：同一个字典类型只能存一行，表实际上不可用。
-- 本迁移删除该约束。必须对存量库幂等，因此对约束名做存在性判断。
--
-- 从 000001 新建的库没有该约束，本迁移为空操作。

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'sys_dict'
          AND c.contype = 'u'
          AND c.conname = 'sys_dict_dict_type_key'
    ) THEN
        ALTER TABLE sys_dict DROP CONSTRAINT sys_dict_dict_type_key;
    END IF;
END $$;

-- 复合唯一约束 uq_sys_dict_type_code 由 000001 创建；
-- 存量库若缺失（例如手工建表）则补上，保证语义一致。
DO $$
BEGIN
    IF to_regclass('public.sys_dict') IS NULL THEN
        RETURN;
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'sys_dict'
          AND c.conname = 'uq_sys_dict_type_code'
    ) THEN
        ALTER TABLE sys_dict
            ADD CONSTRAINT uq_sys_dict_type_code UNIQUE (dict_type, code);
    END IF;
END $$;
