package media_test

import (
	"os"
	"path/filepath"
	"testing"

	"videoCluster/internal/node/media"
)

func TestFileFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "video.bin")
	content := make([]byte, 8<<20)
	for i := range content {
		content[i] = byte(i % 251)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	original := fingerprintForTest(t, path)
	if got := fingerprintForTest(t, path); got != original {
		t.Fatalf("fingerprint is not deterministic: got %q, want %q", got, original)
	}

	tests := []struct {
		name   string
		offset int
	}{
		{name: "beginning", offset: 0},
		{name: "middle", offset: 4 << 20},
		{name: "end", offset: len(content) - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := append([]byte(nil), content...)
			changed[tt.offset]++
			if err := os.WriteFile(path, changed, 0o600); err != nil {
				t.Fatal(err)
			}
			if got := fingerprintForTest(t, path); got == original {
				t.Fatalf("fingerprint did not detect a change in the %s sample", tt.name)
			}
		})
	}
}

func TestFileFingerprintIncludesSize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "video.bin")
	if err := os.WriteFile(path, make([]byte, 4<<20), 0o600); err != nil {
		t.Fatal(err)
	}
	original := fingerprintForTest(t, path)

	if err := os.Truncate(path, (4<<20)+1); err != nil {
		t.Fatal(err)
	}
	if got := fingerprintForTest(t, path); got == original {
		t.Fatal("fingerprint did not detect a file size change")
	}
}

func fingerprintForTest(t *testing.T, path string) string {
	t.Helper()
	fingerprint, err := media.FileFingerprint(path)
	if err != nil {
		t.Fatal(err)
	}
	return fingerprint
}
