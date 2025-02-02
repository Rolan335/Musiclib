package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rolan335/Musiclib/internal/entity"
	"github.com/Rolan335/Musiclib/internal/logger"
)

type Config struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     int    `env:"POSTGRES_PORT"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Name     string `env:"POSTGRES_NAME"`
}

type Storage struct {
	db *pgxpool.Pool
	l  *logger.Log
}

type Song struct {
	ID          int
	Group       string
	Song        string
	ReleaseDate time.Duration
	Text        string
	Link        string
}

func MustNewStorage(cfg *Config, l *logger.Log) *Storage {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	conn, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}

	if err := conn.Ping(context.Background()); err != nil {
		panic("can't connect to postgres: " + err.Error())
	}
	return &Storage{
		db: conn,
		l:  l,
	}
}

func (s *Storage) SelectSongs(ctx context.Context, params entity.GetSongsParams) (song []entity.Song, err error) {
	defer func() {
		s.l.Standart(ctx, "postgres: Select Songs", s.l.FormatGetSongParams(params), song, err)
	}()
	var buf strings.Builder
	//initial query
	buf.WriteString(`SELECT id, "group", title, release_date, text, link FROM songs WHERE 1=1`)
	args := make([]interface{}, 0, 7) // total count of params is 7, to avoid reallocation
	//less than 9 params (7) so we can use runes for indexes to concat faster
	index := '1'

	if params.Group != nil {
		buf.WriteString(` AND "group" = $`)
		buf.WriteRune(index)
		index++
		args = append(args, *params.Group)
	}

	if params.Title != nil {
		buf.WriteString(" AND title = $")
		buf.WriteRune(index)
		index++
		args = append(args, *params.Title)
	}

	if params.Text != nil {
		buf.WriteString(" AND text LIKE $")
		buf.WriteRune(index)
		index++
		args = append(args, "%"+*params.Text+"%")
	}

	if params.DateFrom != nil {
		buf.WriteString(" AND release_date >= $")
		buf.WriteRune(index)
		index++
		args = append(args, *params.DateFrom)
	}

	if params.DateTo != nil {
		buf.WriteString(" AND release_date <= $")
		buf.WriteRune(index)
		index++
		args = append(args, *params.DateTo)
	}
	buf.WriteString(" ORDER BY id ASC")
	//pagination
	if params.Page != nil && params.PageSize != nil {
		buf.WriteString(" LIMIT $")
		buf.WriteRune(index)
		index++
		buf.WriteString(" OFFSET $")
		buf.WriteRune(index)
		args = append(args, *params.PageSize, (*params.Page-1)*(*params.PageSize))
	}
	rows, err := s.db.Query(ctx, buf.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select: %w", err)
	}
	defer rows.Close()
	songs := make([]entity.Song, 0)
	for rows.Next() {
		var song entity.Song
		err := rows.Scan(&song.ID, &song.Group, &song.Title, &song.ReleaseDate, &song.Text, &song.Link)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		songs = append(songs, song)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error in row: %w", rows.Err())
	}

	return songs, nil
}

func (s *Storage) CreateSong(ctx context.Context, song entity.Song) (ID int, err error) {
	defer func() {
		s.l.Standart(ctx, "postgres: CreateSong", song, ID, err)
	}()
	query := `INSERT INTO songs ("group", title, release_date, text, link) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	if err := s.db.QueryRow(ctx, query, song.Group, song.Title, song.ReleaseDate, song.Text, song.Link).
		Scan(&ID); err != nil {
		return 0, fmt.Errorf("failed to exec insert: %w", err)
	}
	return ID, nil
}

func (s *Storage) DeleteSong(ctx context.Context, id int) (err error) {
	defer func() {
		if errors.Is(err, ErrNotFound) {
			s.l.BadInput(ctx, "postgres: DeleteSong", id, err)
			return
		}
		s.l.Standart(ctx, "postgres: DeleteSong", id, nil, err)
	}()
	query := `DELETE FROM songs WHERE id = $1`
	res, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to exec delete: %w", err)
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("data with provided id not found: %w", ErrNotFound)
	}
	return nil
}

func (s *Storage) UpdateSong(ctx context.Context, id int, song entity.SongNullable) (err error) {
	defer func() {
		params := map[string]interface{}{
			"id":   id,
			"song": s.l.FormatSongNullable(song),
		}
		if errors.Is(err, ErrNotFound) {
			s.l.BadInput(ctx, "postgres: UpdateSong", params, err)
			return
		}
		s.l.Standart(ctx, "postgres: UpdateSong", params, nil, err)
	}()
	//Check if song exists
	if err := s.db.QueryRow(ctx, "SELECT id FROM songs WHERE id = $1 LIMIT 1", id).Scan(nil); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("data with provided id not found: %w", ErrNotFound)
		}
		return fmt.Errorf("failed to select song id: %w", err)
	}

	var buf strings.Builder
	//initial query
	buf.WriteString(`UPDATE songs SET`)
	args := make([]interface{}, 0, 5) // total count of params is 5, to avoid reallocation
	//less than 9 params (5) so we can use runes for indexes to concat faster
	index := '1'
	comma := ','

	if song.Group != nil {
		buf.WriteString(` "group" = $`)
		buf.WriteRune(index)
		buf.WriteRune(comma)
		index++
		args = append(args, *song.Group)
	}

	if song.Title != nil {
		buf.WriteString(" title = $")
		buf.WriteRune(index)
		buf.WriteRune(comma)
		index++
		args = append(args, *song.Title)
	}

	if song.ReleaseDate != nil {
		buf.WriteString(" release_date = $")
		buf.WriteRune(index)
		buf.WriteRune(comma)
		index++
		args = append(args, *song.ReleaseDate)
	}

	if song.Text != nil {
		buf.WriteString(" text = $")
		buf.WriteRune(index)
		buf.WriteRune(comma)
		index++
		args = append(args, *song.Text)
	}

	if song.Link != nil {
		buf.WriteString(" link = $")
		buf.WriteRune(index)
		buf.WriteRune(comma)
		index++
		args = append(args, *song.Link)
	}

	//delete last comma and add id
	query := buf.String()
	query = query[:len(query)-1] + " WHERE id = $" + string(index)
	args = append(args, id)
	fmt.Println(query)
	fmt.Println(args...)
	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	return nil
}

func (s *Storage) GetSong(ctx context.Context, id int) (song entity.Song, err error) {
	defer func() {
		if errors.Is(err, ErrNotFound) {
			s.l.BadInput(ctx, "postgres: GetSong", id, err)
			return
		}
		s.l.Standart(ctx, "postgres: GetSong", id, song, err)
	}()
	query := `SELECT "group", title, release_date, text, link FROM songs WHERE id = $1 LIMIT 1`
	if err := s.db.QueryRow(ctx, query, id).Scan(&song.Group, &song.Title, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Song{}, fmt.Errorf("data with provided id not found: %w", ErrNotFound)
		}
		return entity.Song{}, fmt.Errorf("failed to select song: %w", err)
	}
	return song, nil
}

func (s *Storage) Close() {
	s.db.Close()
}
