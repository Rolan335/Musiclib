package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Rolan335/Musiclib/internal/entity"
	"github.com/Rolan335/Musiclib/internal/musiclib"
	"github.com/Rolan335/Musiclib/pkg/api"
	"github.com/Rolan335/Musiclib/pkg/musicinfo"
	"github.com/gin-gonic/gin"
)

type Server struct {
	timeout      time.Duration
	extApiClient *musicinfo.Client
	service      *musiclib.MusicLib
}

func MustNewServer(service *musiclib.MusicLib, externalApiURL string, timeout time.Duration) *Server {
	//Creating client for external api
	client, err := musicinfo.NewClient(externalApiURL)
	if err != nil {
		panic("can't create client: " + err.Error())
	}
	return &Server{
		extApiClient: client,
		service:      service,
		timeout:      timeout,
	}
}

// GetSongs godoc
// @Summary Получение данных библиотеки с фильтрацией по всем полям и пагинацией
// @Description Получение песен с фильтрацией по имени исполнителя, названию песни, тексту песни и диапазону дат с поддержкой пагинации
// @ID get-songs
// @Accept json
// @Produce json
// @Param group query string false "Фильтрация по имени исполнителя"
// @Param title query string false "Фильтрация по названию песни"
// @Param text query string false "Поиск по тексту песни"
// @Param date_from query string false "Фильтрация — песни после указанной даты (YYYY-MM-DD)"
// @Param date_to query string false "Фильтрация — песни до указанной даты (YYYY-MM-DD)"
// @Param page query int false "Номер страницы"
// @Param page_size query int false "Количество элементов на странице"
// @Success 200 {array} entity.Song "Library data"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /songs [get]
func (s *Server) GetSongs(c *gin.Context, params api.GetSongsParams) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), s.timeout)
	defer cancel()
	//Parsing date params
	var dateFrom, dateTo *time.Time
	if params.DateFrom != nil {
		dateFrom = &params.DateFrom.Time
	}
	if params.DateTo != nil {
		dateTo = &params.DateTo.Time
	}
	songs, err := s.service.GetSongs(ctx, entity.GetSongsParams{
		Group:    params.Group,
		Title:    params.Title,
		Text:     params.Text,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": ErrInternalServer.Error()})
		return
	}
	c.JSON(200, songs)
}

// PostSongs godoc
// @Summary Добавление новой песни
// @Description Добавление новой песни с извлечением информации о песне из внешнего API
// @Accept json
// @Produce json
// @Param song body Song true "Song data"
// @Success 201 {object} map[string]int "Successfully added"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 500 {object} gin.H "Internal server error"
// @Failure 504 {object} gin.H "Gateway timeout"
// @Router /songs [post]
func (s *Server) PostSongs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), s.timeout)
	defer cancel()
	var song Song
	if err := c.BindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrFailedToParse.Error()})
		return
	}

	resp, err := s.extApiClient.GetInfo(ctx, &musicinfo.GetInfoParams{Song: song.Title, Group: song.Group})
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": ErrGatewayTimeout.Error()})
		return
	}
	defer resp.Body.Close()
	//Если во внешнем апи не нашлось информации о песне, то не добавляем ее в базу
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound.Error()})
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}
	if err := json.Unmarshal(body, &song); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrFailedToParse.Error()})
		return
	}
	id, err := s.service.CreateSong(ctx, entity.Song{
		Group:       song.Group,
		Title:       song.Title,
		ReleaseDate: song.ReleaseDate.Time(),
		Text:        song.Text,
		Link:        song.Link,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// DeleteSongsId godoc
// @Summary Удаление песни
// @Description Удаление песни по её ID
// @Param id path int true "Song ID"
// @Success 204 {object} gin.H "Successfully deleted"
// @Failure 404 {object} gin.H "Song not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /songs/{id} [delete]
func (s *Server) DeleteSongsId(c *gin.Context, id int) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), s.timeout)
	defer cancel()
	if err := s.service.DeleteSong(ctx, id); err != nil {
		if errors.Is(err, musiclib.ErrSongNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// PatchSongsId godoc
// @Summary Изменение данных песни
// @Description Обновление данных песни по её ID
// @Param id path int true "Song ID"
// @Param song body SongNullable true "Updated song data"
// @Success 204 {object} gin.H "Successfully updated"
// @Failure 404 {object} gin.H "Song not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /songs/{id} [patch]
func (s *Server) PatchSongsId(c *gin.Context, id int) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), s.timeout)
	defer cancel()
	var song SongNullable
	if err := c.BindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrFailedToParse.Error()})
		return
	}
	//Check for nil time
	var releaseDate *time.Time
	if song.ReleaseDate != nil {
		time := song.ReleaseDate.Time()
		releaseDate = &time
	}
	err := s.service.UpdateSong(ctx, id, entity.SongNullable{
		ID:          &id,
		Group:       song.Group,
		Title:       song.Title,
		ReleaseDate: releaseDate,
		Text:        song.Text,
		Link:        song.Link,
	})
	if err != nil {
		if errors.Is(err, musiclib.ErrSongNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) GetSongsIdText(c *gin.Context, id int, params api.GetSongsIdTextParams) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), s.timeout)
	defer cancel()
	//default values if nil
	page := 1
	pageSize := 1
	if params.Page != nil {
		page = *params.Page
	}
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	text, err := s.service.GetSongText(ctx, id, page, pageSize)
	if err != nil {
		if errors.Is(err, musiclib.ErrSongNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound.Error()})
			return
		}
		if errors.Is(err, musiclib.ErrInvalidParams) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrBadRequest.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}
	c.JSON(http.StatusOK, text)
}
