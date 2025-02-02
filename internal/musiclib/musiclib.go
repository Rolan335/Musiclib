// Business logic
package musiclib

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Rolan335/Musiclib/internal/entity"
	"github.com/Rolan335/Musiclib/internal/logger"
	"github.com/Rolan335/Musiclib/internal/repository/postgres"
)

type Storage interface {
	SelectSongs(ctx context.Context, params entity.GetSongsParams) ([]entity.Song, error)
	CreateSong(ctx context.Context, song entity.Song) (int, error)
	DeleteSong(ctx context.Context, id int) error
	UpdateSong(ctx context.Context, id int, song entity.SongNullable) error
	GetSong(ctx context.Context, id int) (entity.Song, error)
}

type MusicLib struct {
	storage Storage
	log     *logger.Log
}

func NewMusicLib(storage Storage, l *logger.Log) *MusicLib {
	return &MusicLib{
		storage: storage,
		log:     l,
	}
}

func (m *MusicLib) GetSongs(ctx context.Context, params entity.GetSongsParams) (songs []entity.Song, err error) {
	defer func() {
		m.log.Standart(ctx, "musiclib: GetSongs", m.log.FormatGetSongParams(params), songs, err)
	}()
	//default values for page and pageSize
	if params.Page == nil {
		params.Page = new(int)
		*params.Page = 1
	}
	if params.PageSize == nil {
		params.PageSize = new(int)
		*params.PageSize = 10
	}
	songs, err = m.storage.SelectSongs(ctx, params)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

// returns id
func (m *MusicLib) CreateSong(ctx context.Context, song entity.Song) (songID int, err error) {
	defer func() {
		m.log.Standart(ctx, "musiclib: CreateSong", song, songID, err)
	}()
	songID, err = m.storage.CreateSong(ctx, song)
	if err != nil {
		return 0, fmt.Errorf("failed to create song: %w", err)
	}
	return songID, nil
}

func (m *MusicLib) DeleteSong(ctx context.Context, id int) (err error) {
	defer func() {
		if errors.Is(err, ErrSongNotFound) {
			m.log.BadInput(ctx, "musiclib: DeleteSong", id, err)
			return
		}
		m.log.Standart(ctx, "musiclib: DeleteSong", id, nil, err)
	}()
	err = m.storage.DeleteSong(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return fmt.Errorf("db didn't find song with id %d: %w", id, ErrSongNotFound)
		}
		return fmt.Errorf("db error: %w", err)
	}
	return nil
}

func (m *MusicLib) UpdateSong(ctx context.Context, id int, song entity.SongNullable) (err error) {
	defer func() {
		if errors.Is(err, ErrSongNotFound) {
			m.log.BadInput(ctx, "musiclib: UpdateSong", m.log.FormatSongNullable(song), err)
			return
		}
		m.log.Standart(ctx, "musiclib: UpdateSong", m.log.FormatSongNullable(song), nil, err)
	}()
	err = m.storage.UpdateSong(ctx, id, song)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return fmt.Errorf("db didn't find song with id %d: %w", id, ErrSongNotFound)
		}
		return fmt.Errorf("db error: %w", err)
	}
	return nil
}

func (m *MusicLib) GetSongText(ctx context.Context, id int, page int, pageSize int) (text entity.Text, err error) {
	defer func() {
		params := map[string]int{
			"id":       id,
			"page":     page,
			"pageSize": pageSize,
		}
		if errors.Is(err, ErrSongNotFound) || errors.Is(err, ErrInvalidParams) {
			m.log.BadInput(ctx, "musiclib: GetSongText", params, err)
			return
		}
		m.log.Standart(ctx, "musiclib: GetSongText", params, text, err)
	}()
	song, err := m.storage.GetSong(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return entity.Text{}, fmt.Errorf("db didn't find song with id %d: %w", id, ErrSongNotFound)
		}
		return entity.Text{}, fmt.Errorf("db error: %w", err)
	}
	//Pagination, split verses by \n\n
	//page - from which verse start (offset)
	//pageSize - how much verses to take (limit)
	//example - page=1 pageSize=2 - returns 2 verses from 1 included
	verses := strings.Split(song.Text, "\n\n")
	if page > len(verses) {
		return entity.Text{}, fmt.Errorf("page is bigger than number of verses: %w", ErrInvalidParams)
	}
	startIndex := page - 1
	endIndex := startIndex + pageSize
	if endIndex > len(verses) {
		endIndex = len(verses)
	}
	newSlice := make([][]string, len(verses[startIndex:endIndex]))
	fmt.Println(len(newSlice))
	for i := range newSlice {
		newSlice[i] = strings.Split(verses[startIndex:endIndex][i], "\n")
	}
	text = entity.Text{
		Text: newSlice,
	}
	return text, nil
}
