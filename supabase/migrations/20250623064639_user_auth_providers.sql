create type auth_provider as enum ('clerk');

ALTER TABLE users
ADD COLUMN auth_provider auth_provider DEFAULT 'clerk',
ADD COLUMN auth_provider_id text;