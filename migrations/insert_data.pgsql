-- Вставляем исполнителей
INSERT INTO artists (artist_name)
VALUES 
    ('Muse'),
    ('The Beatles'),
    ('Queen'),
    ('Led Zeppelin'),
    ('Pink Floyd'),
    ('Nirvana');

-- Вставляем песни, связывая их с исполнителями
INSERT INTO songs (artist_id, title, release_date, lyrics, link)
SELECT 
    a.artist_id,
    s.title,
    s.release_date::date,
    s.lyrics,
    s.link
FROM (
    VALUES 
        ('Muse', 'Supermassive Black Hole', '2006-07-16', 'Ooh baby, don’t you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight', 'https://www.youtube.com/watch?v=Xsp3_a-PMTw'),
        ('The Beatles', 'Hey Jude', '1968-08-26', 'Hey Jude, don’t make it bad.\nTake a sad song and make it better.\nRemember to let her into your heart,\nThen you can start to make it better.\n\nHey Jude, don’t be afraid.\nYou were made to go out and get her.\nThe minute you let her under your skin,\nThen you begin to make it better.\n\nAnd anytime you feel the pain, hey Jude, refrain', 'http://example.com/heyjude'),
        ('Queen', 'Bohemian Rhapsody', '1975-10-31', 'Is this the real life? Is this just fantasy? Caught in a landslide, no escape from reality...', 'http://example.com/bohemianrhapsody'),
        ('Led Zeppelin', 'Stairway to Heaven', '1971-11-08', 'There’s a lady who’s sure all that glitters is gold, and she’s buying a stairway to heaven...', 'http://example.com/stairwaytoheaven'),
        ('Pink Floyd', 'Wish You Were Here', '1975-09-12', 'So, so you think you can tell Heaven from Hell, blue skies from pain...', 'http://example.com/wishyouwerehere'),
        ('Nirvana', 'Smells Like Teen Spirit', '1991-09-10', 'Load up on guns, bring your friends. It’s fun to lose and to pretend...', 'http://example.com/smellsliketeenspirit')
    ) AS s(artist_name, title, release_date, lyrics, link)
JOIN artists a ON a.artist_name = s.artist_name;