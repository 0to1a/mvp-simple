-- name: GetUserByID :one
SELECT id, email, name, created_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT 
    u.id,
    u.email,
    u.name,
    u.created_at,
    uc.is_admin
FROM users u
JOIN user_companies uc ON u.id = uc.user_id
WHERE uc.company_id = $1 AND u.deleted_at IS NULL
ORDER BY u.created_at DESC;

-- name: CreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING id, email, name, created_at;

-- name: AddUserToCompany :exec
INSERT INTO user_companies (user_id, company_id, is_admin)
VALUES ($1, $2, $3);

-- name: CheckUserInCompany :one
SELECT EXISTS (
    SELECT 1 
    FROM user_companies 
    WHERE user_id = $1 AND company_id = $2
);