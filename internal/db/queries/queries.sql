-- name: GetUserByEmail :one
SELECT id, email, name, created_at
FROM users
WHERE email = $1;

-- name: GetUserCompanies :many
SELECT 
    c.id as company_id,
    c.name as company_name,
    uc.is_admin,
    uc.created_at
FROM user_companies uc
JOIN companies c ON c.id = uc.company_id
WHERE uc.user_id = $1
ORDER BY uc.created_at ASC;

-- name: GetDefaultUserCompany :one
SELECT 
    c.id as company_id,
    c.name as company_name,
    uc.is_admin
FROM user_companies uc
JOIN companies c ON c.id = uc.company_id
WHERE uc.user_id = $1
ORDER BY uc.created_at ASC
LIMIT 1;
