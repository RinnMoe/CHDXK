UPDATE users
SET role = 'admin'
WHERE role = 'super_admin';

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
    ADD CONSTRAINT users_role_check
        CHECK (role IN ('user', 'admin'));
