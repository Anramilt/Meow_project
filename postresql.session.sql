CREATE TABLE sounds (
 id SERIAL PRIMARY KEY,
 id_game INTEGER REFERENCES games(id_game) ON DELETE CASCADE,
 sound_name TEXT NOT NULL,
 UNIQUE(id_game, sound_name)
);
