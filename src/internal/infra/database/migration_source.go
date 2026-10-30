package database

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

var singleFileMigrationRegex = regexp.MustCompile(`^([0-9]+)_([a-zA-Z0-9_\-]+)\.sql$`)

type migrationEntry struct {
	version    uint
	identifier string
	upSQL      string
	downSQL    string
}

// SingleFileSource adalah implementasi source.Driver untuk golang-migrate
// yang membaca 1 file .sql berisi blok "-- +migrate Up" dan "-- +migrate Down".
type SingleFileSource struct {
	path       string
	migrations *source.Migrations
	entries    map[uint]migrationEntry
}

// NewSingleFileSource memindai direktori migrasi dan mem-parse setiap file <version>_<name>.sql.
func NewSingleFileSource(dirPath string) (source.Driver, error) {
	dir := strings.TrimPrefix(strings.TrimSpace(dirPath), "file://")
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat migrations dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migrations path %q is not a directory", dir)
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	migrations := source.NewMigrations()
	entries := make(map[uint]migrationEntry)

	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}

		matches := singleFileMigrationRegex.FindStringSubmatch(de.Name())
		if len(matches) != 3 {
			continue
		}

		v64, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse version in %q: %w", de.Name(), err)
		}
		version := uint(v64)
		identifier := matches[2]

		fullPath := filepath.Join(dir, de.Name())
		rawBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("read migration file %q: %w", fullPath, err)
		}

		upSQL, downSQL, err := ParseSingleFileSQL(string(rawBytes))
		if err != nil {
			return nil, fmt.Errorf("parse migration file %q: %w", de.Name(), err)
		}

		entries[version] = migrationEntry{
			version:    version,
			identifier: identifier,
			upSQL:      upSQL,
			downSQL:    downSQL,
		}

		upMig := &source.Migration{
			Version:    version,
			Identifier: identifier,
			Direction:  source.Up,
			Raw:        de.Name(),
		}
		if !migrations.Append(upMig) {
			return nil, fmt.Errorf("duplicate migration version %d in %q", version, de.Name())
		}

		downMig := &source.Migration{
			Version:    version,
			Identifier: identifier,
			Direction:  source.Down,
			Raw:        de.Name(),
		}
		if !migrations.Append(downMig) {
			return nil, fmt.Errorf("duplicate down migration version %d in %q", version, de.Name())
		}
	}

	return &SingleFileSource{
		path:       dir,
		migrations: migrations,
		entries:    entries,
	}, nil
}

// ParseSingleFileSQL memisahkan isi file .sql menjadi query Up dan Down
// berdasarkan marker "-- +migrate Up" dan "-- +migrate Down".
func ParseSingleFileSQL(content string) (upSQL string, downSQL string, err error) {
	const (
		sectionNone = iota
		sectionUp
		sectionDown
	)

	var (
		currentSection = sectionNone
		upBuf          strings.Builder
		downBuf        strings.Builder
		hasUpMarker    bool
		hasDownMarker  bool
	)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.ToLower(strings.TrimSpace(line))

		switch trimmed {
		case "-- +migrate up", "-- migrate:up", "-- up":
			currentSection = sectionUp
			hasUpMarker = true
			continue
		case "-- +migrate down", "-- migrate:down", "-- down":
			currentSection = sectionDown
			hasDownMarker = true
			continue
		}

		switch currentSection {
		case sectionUp:
			upBuf.WriteString(line)
			upBuf.WriteByte('\n')
		case sectionDown:
			downBuf.WriteString(line)
			downBuf.WriteByte('\n')
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	if !hasUpMarker {
		return "", "", fmt.Errorf("missing '-- +migrate Up' marker")
	}
	if !hasDownMarker {
		return "", "", fmt.Errorf("missing '-- +migrate Down' marker")
	}

	return strings.TrimSpace(upBuf.String()), strings.TrimSpace(downBuf.String()), nil
}

func (s *SingleFileSource) Open(url string) (source.Driver, error) {
	return NewSingleFileSource(url)
}

func (s *SingleFileSource) Close() error {
	return nil
}

func (s *SingleFileSource) First() (version uint, err error) {
	v, ok := s.migrations.First()
	if !ok {
		return 0, &os.PathError{Op: "first", Path: s.path, Err: os.ErrNotExist}
	}
	return v, nil
}

func (s *SingleFileSource) Prev(version uint) (prevVersion uint, err error) {
	v, ok := s.migrations.Prev(version)
	if !ok {
		return 0, &os.PathError{Op: fmt.Sprintf("prev for version %v", version), Path: s.path, Err: os.ErrNotExist}
	}
	return v, nil
}

func (s *SingleFileSource) Next(version uint) (nextVersion uint, err error) {
	v, ok := s.migrations.Next(version)
	if !ok {
		return 0, &os.PathError{Op: fmt.Sprintf("next for version %v", version), Path: s.path, Err: os.ErrNotExist}
	}
	return v, nil
}

func (s *SingleFileSource) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	entry, ok := s.entries[version]
	if !ok {
		return nil, "", &os.PathError{Op: fmt.Sprintf("read up for version %v", version), Path: s.path, Err: os.ErrNotExist}
	}
	return io.NopCloser(bytes.NewReader([]byte(entry.upSQL))), entry.identifier, nil
}

func (s *SingleFileSource) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	entry, ok := s.entries[version]
	if !ok {
		return nil, "", &os.PathError{Op: fmt.Sprintf("read down for version %v", version), Path: s.path, Err: os.ErrNotExist}
	}
	return io.NopCloser(bytes.NewReader([]byte(entry.downSQL))), entry.identifier, nil
}

