package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dobby/filemanager/internal/application"
)

// OSFileSystem implements application.IFileSystem using the standard os package.
type OSFileSystem struct{}

func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ListFiles returns all files (not directories) under dir.
// When recursive is true, subdirectories are traversed as well.
func (f *OSFileSystem) ListFiles(_ context.Context, dir string, recursive bool) ([]application.FileInfo, error) {
	var results []application.FileInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if recursive {
				sub, err := f.ListFiles(context.Background(), fullPath, true)
				if err != nil {
					continue
				}
				results = append(results, sub...)
			}
			continue
		}

		ext := filepath.Ext(entry.Name())
		name := strings.TrimSuffix(entry.Name(), ext)

		info, err := entry.Info()
		var detectedAt time.Time
		if err == nil {
			detectedAt = info.ModTime()
		} else {
			detectedAt = time.Now()
		}

		results = append(results, application.FileInfo{
			Path:       fullPath,
			Name:       name,
			Extension:  ext,
			DetectedAt: detectedAt,
		})
	}

	return results, nil
}

// MoveFile moves (and renames) the file at src to dst.
func (f *OSFileSystem) MoveFile(_ context.Context, src, dst string) error {
	return os.Rename(src, dst)
}

// EnsureDir creates dir and all necessary parents.
func (f *OSFileSystem) EnsureDir(_ context.Context, dir string) error {
	return os.MkdirAll(dir, 0o755)
}
