CREATE TABLE
    role_permissions (
        id CHAR(36) NOT NULL PRIMARY KEY,
        role_id CHAR(36) NOT NULL,
        permission_id CHAR(36) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        created_by CHAR(36) NOT NULL,
        deleted_at TIMESTAMP NULL,
        -- Foreign key constraints
        CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles (id) ON UPDATE CASCADE ON DELETE CASCADE,
        CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions (id) ON UPDATE CASCADE ON DELETE CASCADE,
        -- Indexes
        INDEX idx_role_permissions_role_id (role_id),
        INDEX idx_role_permissions_permission_id (permission_id),
        INDEX idx_role_permissions_deleted_at (deleted_at),
        -- Unique constraint to prevent duplicate role-permission combinations
        UNIQUE INDEX idx_role_permissions_unique (role_id, permission_id, deleted_at)
    );