SELECT DISTINCT category.tag 
FROM category
JOIN accordance_game_category ON category.id_category = accordance_game_category.id_category
JOIN game ON accordance_game_category.id_game = game.id_game
WHERE game.name_game = 'Большой-маленький';
