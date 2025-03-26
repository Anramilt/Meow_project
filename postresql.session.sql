SELECT DISTINCT name_game
            FROM games
            WHERE levenshtein(name_game, 'Автоматизация звука') <= 8
            LIMIT 3;
