package loader

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFile(t *testing.T) {
	// Setup temporary directory and file
	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "valid.txt")
	content := []byte("hello world")
	if err := os.WriteFile(validFile, content, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Load existing file",
			args:    args{path: validFile},
			want:    content,
			wantErr: false,
		},
		{
			name:    "Load non-existent file",
			args:    args{path: filepath.Join(tempDir, "nonexistent.txt")},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(tempFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "Path is a directory",
			args:    args{path: tempDir},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Path is a file",
			args:    args{path: tempFile},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Path does not exist",
			args:    args{path: filepath.Join(tempDir, "nonexistent")},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsDir(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListYAMLFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create mixed files
	files := map[string]string{
		"config.yaml": "data: 1",
		"script.sh":   "echo hello",
		"app.yml":     "version: 1",
		"notes.txt":   "readme",
	}

	for name, content := range files {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	// Create subdirectory (should be ignored)
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "ignored.yaml"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create ignored file: %v", err)
	}

	// Create empty dir
	emptyDir := t.TempDir()

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "List YAML files in mixed directory",
			args:    args{dir: tempDir},
			want:    []string{"app.yml", "config.yaml"}, // Sorted alphabetically
			wantErr: false,
		},
		{
			name:    "List YAML files in empty directory",
			args:    args{dir: emptyDir},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "List YAML files in non-existent directory",
			args:    args{dir: filepath.Join(tempDir, "nonexistent_dir")},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListYAMLFiles(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListYAMLFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListYAMLFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
