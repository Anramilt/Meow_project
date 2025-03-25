CREATE TABLE images (
   id SERIAL PRIMARY KEY,
   id_game INT REFERENCES games(id_game) ON DELETE CASCADE,
   image_name TEXT NOT NULL,
   is_correct BOOLEAN
);
