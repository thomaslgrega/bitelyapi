ALTER TABLE recipes
ADD COLUMN user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX recipes_user_id_idx ON recipes(user_id);