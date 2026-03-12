-- Hermes Platform 数据库初始化脚本
-- 版本: 1.0.0
-- 日期: 2026-03-10

-- 1. 创建扩展（如果不存在）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. 删除现有表（如果存在）
DROP TABLE IF EXISTS test_records;
DROP TABLE IF EXISTS test_details;
DROP TABLE IF EXISTS test_tasks;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;

-- 3. 创建表结构

-- 用户表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 角色表
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 权限表
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    code VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 测试任务表
CREATE TABLE test_tasks (
    id SERIAL PRIMARY KEY,
    task_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_time BIGINT NOT NULL,
    end_time BIGINT,
    duration BIGINT,
    total_tests INTEGER DEFAULT 0,
    passed_tests INTEGER DEFAULT 0,
    failed_tests INTEGER DEFAULT 0,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 测试详情表
CREATE TABLE test_details (
    id SERIAL PRIMARY KEY,
    test_task_id INTEGER NOT NULL,
    test_name VARCHAR(255) NOT NULL,
    test_status VARCHAR(50) NOT NULL,
    error_message TEXT,
    execution_time BIGINT,
    test_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (test_task_id) REFERENCES test_tasks(id) ON DELETE CASCADE
);

-- 测试记录表
CREATE TABLE test_records (
    id SERIAL PRIMARY KEY,
    test_task_id INTEGER NOT NULL,
    test_name VARCHAR(255) NOT NULL,
    record_type VARCHAR(50) NOT NULL,
    record_data TEXT,
    record_time BIGINT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (test_task_id) REFERENCES test_tasks(id) ON DELETE CASCADE
);

-- 多对多关系表
CREATE TABLE user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- 4. 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_test_tasks_status ON test_tasks(status);
CREATE INDEX IF NOT EXISTS idx_test_tasks_start_time ON test_tasks(start_time);
CREATE INDEX IF NOT EXISTS idx_test_tasks_user_id ON test_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_test_details_test_task_id ON test_details(test_task_id);
CREATE INDEX IF NOT EXISTS idx_test_details_test_status ON test_details(test_status);
CREATE INDEX IF NOT EXISTS idx_test_records_test_task_id ON test_records(test_task_id);
CREATE INDEX IF NOT EXISTS idx_test_records_record_type ON test_records(record_type);
CREATE INDEX IF NOT EXISTS idx_test_records_record_time ON test_records(record_time);

-- 5. 插入默认数据

-- 插入默认权限
INSERT INTO permissions (name, description, code) VALUES
('用户管理', '管理用户账户', 'user.manage'),
('角色管理', '管理角色', 'role.manage'),
('权限管理', '管理权限', 'permission.manage'),
('测试任务管理', '管理测试任务', 'test.task.manage'),
('测试详情管理', '管理测试详情', 'test.detail.manage'),
('测试记录管理', '管理测试记录', 'test.record.manage'),
('系统设置', '管理系统设置', 'system.settings'),
('数据导出', '导出数据', 'data.export'),
('数据导入', '导入数据', 'data.import');

-- 插入默认角色
INSERT INTO roles (name, description) VALUES
('管理员', '系统管理员，拥有所有权限'),
('普通用户', '普通用户，拥有基本权限');

-- 为角色分配权限
-- 管理员角色分配所有权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8), (1, 9);

-- 普通用户角色分配基本权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(2, 4), (2, 5), (2, 6), (2, 8);

-- 插入默认管理员用户
-- 密码：admin123（哈希值）
INSERT INTO users (name, email, password) VALUES
('Admin', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy');

-- 为管理员用户分配管理员角色
INSERT INTO user_roles (user_id, role_id) VALUES
(1, 1);

-- 6. 提交事务
COMMIT;

-- 7. 完成信息
SELECT 'Hermes Platform 数据库初始化完成' AS message;
