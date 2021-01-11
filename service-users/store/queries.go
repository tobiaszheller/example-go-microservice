package store

const (
	queryInsertUser = `
INSERT INTO users(
	id,
	first_name,
	last_name,
	nickname,
	email,
	country,
	updated_at
) VALUES (
	:id,
	:first_name,
	:last_name,
	:nickname,
	:email,
	:country,
	:updated_at
);
`
	queryUpdateUser = `
UPDATE
	users
SET
	first_name = :first_name,
	last_name = :last_name,
	nickname = :nickname,
	email = :email,
	country = :country,
	updated_at = :updated_at
WHERE
   id = :id;
`

	querySelectUserById = `
SELECT
	id,
	first_name,
	last_name,
	nickname,
	email,
	country,
	updated_at
FROM
	users
WHERE
	id = ?;
`
)
