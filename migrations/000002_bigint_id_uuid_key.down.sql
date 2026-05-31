ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_key_key;
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_pkey;
ALTER TABLE refresh_tokens DROP COLUMN id;
ALTER TABLE refresh_tokens RENAME COLUMN key TO id;
ALTER TABLE refresh_tokens ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);
