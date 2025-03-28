CREATE TABLE game_category (
   id_game INT REFERENCES games(id_game) ON DELETE CASCADE,
   id_category INT REFERENCES category(id_category) ON DELETE CASCADE,
   PRIMARY KEY (id_game, id_category)
);




