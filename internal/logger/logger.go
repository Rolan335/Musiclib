package logger

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/Rolan335/Musiclib/internal/entity"
)

type Log struct {
	*slog.Logger
}

func New(logLevel string, out io.Writer) *Log {
	logLevel = strings.ToUpper(logLevel)
	switch logLevel {
	case "DEBUG":
		return &Log{slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	case "INFO":
		return &Log{slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}))}
	default:
		return &Log{slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}))}
	}
}

// Logging errors on lvl error, and if err == nil, log on lvl debug
func (l *Log) Standart(ctx context.Context, methodName string, params interface{}, result interface{}, err error) {
	if err != nil {
		l.LogAttrs(ctx, slog.LevelError, "musiclib: "+methodName,
			slog.Any("params", params),
			slog.String("error", err.Error()),
		)
		return
	}
	l.LogAttrs(ctx, slog.LevelDebug, "musiclib: "+methodName,
		slog.Any("params", params),
		slog.Any("result", result),
	)
}

// Logging if bad input provided on lvl debug
func (l *Log) BadInput(ctx context.Context, methodName string, params interface{}, err error) {
	if err != nil {
		l.LogAttrs(ctx, slog.LevelDebug, "musiclib: "+methodName,
			slog.Any("params", params),
			slog.String("error", err.Error()),
		)
	}
}

// func for formating songNullable for logging on lvl debug
func (l *Log) FormatSongNullable(song entity.SongNullable) map[string]interface{} {
	return map[string]interface{}{
		"ID":          dereferencePointer(song.ID),
		"Group":       dereferencePointer(song.Group),
		"Title":       dereferencePointer(song.Title),
		"ReleaseDate": dereferencePointer(song.ReleaseDate),
		"Text":        dereferencePointer(song.Text),
		"Link":        dereferencePointer(song.Link),
	}
}

// func for formating GetSongParams for logging on lvl debug
func (l *Log) FormatGetSongParams(songParams entity.GetSongsParams) map[string]interface{} {
	return map[string]interface{}{
		"Group":    dereferencePointer(songParams.Group),
		"Title":    dereferencePointer(songParams.Title),
		"Text":     dereferencePointer(songParams.Text),
		"DateFrom": dereferencePointer(songParams.DateFrom),
		"DateTo":   dereferencePointer(songParams.DateTo),
		"Page":     dereferencePointer(songParams.Page),
		"PageSize": dereferencePointer(songParams.PageSize),
	}
}

func dereferencePointer[T any](ptr *T) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}
