-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS songs(
    ID SERIAL PRIMARY KEY NOT NULL,
    "group" VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    release_date DATE NOT NULL,
    text TEXT NOT NULL,
    link TEXT NOT NULL
);

CREATE INDEX idx_songs_group on songs USING hash ("group");
CREATE INDEX idx_songs_song on songs USING hash (title);
CREATE INDEX idx_songs_release_date on songs (release_date);
CREATE INDEX idx_songs_text_fulltext ON songs USING gin(to_tsvector('english', text));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS songs;
DROP INDEX IF EXISTS idx_songs_release_date;
DROP INDEX IF EXISTS idx_songs_song;
DROP INDEX IF EXISTS idx_songs_group;
-- +goose StatementEnd
