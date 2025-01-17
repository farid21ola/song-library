CREATE TABLE IF NOT EXISTS artists (
   artist_id SERIAL PRIMARY KEY,
   artist_name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS songs (
    song_id SERIAL PRIMARY KEY,
    artist_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    release_date DATE,
    lyrics TEXT,
    link VARCHAR(255),
    UNIQUE (artist_id, title),
    FOREIGN KEY (artist_id) REFERENCES artists(artist_id) ON DELETE CASCADE
);

CREATE INDEX idx_songs_title ON songs(title);
CREATE INDEX idx_songs_release_date ON songs(release_date);