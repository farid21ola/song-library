CREATE TABLE IF NOT EXISTS songs (
    artist VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    release_date DATE,
    lyrics TEXT,
    link VARCHAR(255),
    PRIMARY KEY (artist, title)
);