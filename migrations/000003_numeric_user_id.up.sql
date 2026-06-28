DELETE FROM refresh_tokens;

ALTER TABLE refresh_tokens
    ALTER COLUMN user_id TYPE BIGINT USING NULL;
