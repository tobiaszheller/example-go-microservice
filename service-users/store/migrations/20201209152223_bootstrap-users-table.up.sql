CREATE TABLE users (
  id varchar(36) PRIMARY KEY,
  first_name text NOT NULL,
  last_name text NOT NULL,
  nickname text NOT NULL,
  email varchar(255) NOT NULL UNIQUE,
  country varchar(2) NOT NULL,
  updated_at timestamp NULL DEFAULT NULL
);

CREATE INDEX users_countries ON users (country);
