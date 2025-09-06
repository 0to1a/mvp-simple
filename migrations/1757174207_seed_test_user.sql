-- +goose Up
-- Insert test company
INSERT INTO companies (name) VALUES ('Test Company') ON CONFLICT DO NOTHING;

-- Get company ID and insert test user with admin privileges
INSERT INTO users (email, name)
SELECT 'test@test.com', 'Test User'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'test@test.com');

-- Insert test company and assign test user as admin
INSERT INTO user_companies (user_id, company_id, is_admin)
SELECT u.id, c.id, true
FROM users u, companies c
WHERE u.email = 'test@test.com' AND c.name = 'Test Company'
AND NOT EXISTS (
    SELECT 1 FROM user_companies uc 
    JOIN users u2 ON uc.user_id = u2.id
    JOIN companies c2 ON uc.company_id = c2.id
    WHERE u2.email = 'test@test.com' AND c2.name = 'Test Company'
);

-- +goose Down
-- Remove test user company membership
DELETE FROM user_companies 
WHERE user_id IN (SELECT id FROM users WHERE email = 'test@test.com');

-- Remove test user
DELETE FROM users WHERE email = 'test@test.com';

-- Remove test company if no other users
DELETE FROM companies 
WHERE name = 'Test Company' 
AND NOT EXISTS (SELECT 1 FROM user_companies WHERE company_id = companies.id);

