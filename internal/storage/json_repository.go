package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const formatVersion = 1

var ErrUnsupportedVersion = errors.New("unsupported history file version")

type historyFile struct {
	Version int            `json:"version"`
	Entries []EpisodeEntry `json:"entries"`
}

type JSONRepository struct {
	path string
}

var _ Repository = (*JSONRepository)(nil)

func NewJSONRepository(path string) *JSONRepository {
	return &JSONRepository{path: path}
}

func (r *JSONRepository) SaveHistory(entries []EpisodeEntry) (err error) {
	if entries == nil {
		entries = []EpisodeEntry{}
	}
	data, err := json.MarshalIndent(historyFile{Version: formatVersion, Entries: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding history: %w", err)
	}

	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".history-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing %s: %w", tmp.Name(), err)
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("syncing %s: %w", tmp.Name(), err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", tmp.Name(), err)
	}
	if err = os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("replacing %s: %w", r.path, err)
	}
	return nil
}

func (r *JSONRepository) LoadHistory() ([]EpisodeEntry, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, fs.ErrNotExist) {
		return []EpisodeEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading history file %s: %w", r.path, err)
	}

	var f historyFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing history file %s: %w", r.path, err)
	}
	if f.Version != formatVersion {
		return nil, fmt.Errorf("history file %s has version %d, this program reads version %d: %w",
			r.path, f.Version, formatVersion, ErrUnsupportedVersion)
	}
	if f.Entries == nil {
		f.Entries = []EpisodeEntry{}
	}
	return f.Entries, nil
}
