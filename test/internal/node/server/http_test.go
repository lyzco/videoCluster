package server_test

import (
	"context"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-kratos/kratos/v2/log"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/biz"
	"videoCluster/internal/node/conf"
	"videoCluster/internal/node/data"
	"videoCluster/internal/node/server"
	"videoCluster/internal/node/service"
)

func TestVideoStreaming(t *testing.T) {
	const content = "0123456789"
	httpServer, path := newVideoHTTPServer(t, "video-id", "示例 movie.mkv", []byte(content))

	t.Run("full content", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodGet, path, nil)
		if response.Code != stdhttp.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusOK)
		}
		if response.Body.String() != content {
			t.Fatalf("body = %q, want %q", response.Body.String(), content)
		}
		if response.Header().Get("Accept-Ranges") != "bytes" {
			t.Fatalf("Accept-Ranges = %q, want bytes", response.Header().Get("Accept-Ranges"))
		}
		if response.Header().Get("Content-Type") == "video/mp4" {
			t.Fatal("non-MP4 video was served as video/mp4")
		}
	})

	t.Run("explicit range", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodGet, path, map[string]string{"Range": "bytes=2-5"})
		if response.Code != stdhttp.StatusPartialContent {
			t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusPartialContent)
		}
		if response.Body.String() != "2345" {
			t.Fatalf("body = %q, want %q", response.Body.String(), "2345")
		}
		if got := response.Header().Get("Content-Range"); got != "bytes 2-5/10" {
			t.Fatalf("Content-Range = %q, want %q", got, "bytes 2-5/10")
		}
	})

	t.Run("suffix range", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodGet, path, map[string]string{"Range": "bytes=-4"})
		if response.Code != stdhttp.StatusPartialContent || response.Body.String() != "6789" {
			t.Fatalf("status/body = %d/%q, want %d/%q", response.Code, response.Body.String(), stdhttp.StatusPartialContent, "6789")
		}
	})

	t.Run("multiple ranges", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodGet, path, map[string]string{"Range": "bytes=0-1,8-9"})
		if response.Code != stdhttp.StatusPartialContent {
			t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusPartialContent)
		}
		if !strings.HasPrefix(response.Header().Get("Content-Type"), "multipart/byteranges;") {
			t.Fatalf("Content-Type = %q, want multipart/byteranges", response.Header().Get("Content-Type"))
		}
		if body := response.Body.String(); !strings.Contains(body, "01") || !strings.Contains(body, "89") {
			t.Fatalf("multipart body does not contain both ranges: %q", body)
		}
	})

	t.Run("unsatisfied range", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodGet, path, map[string]string{"Range": "bytes=20-"})
		if response.Code != stdhttp.StatusRequestedRangeNotSatisfiable {
			t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusRequestedRangeNotSatisfiable)
		}
		if got := response.Header().Get("Content-Range"); got != "bytes */10" {
			t.Fatalf("Content-Range = %q, want %q", got, "bytes */10")
		}
	})

	t.Run("head", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodHead, path, nil)
		if response.Code != stdhttp.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusOK)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("HEAD body length = %d, want 0", response.Body.Len())
		}
	})

	t.Run("etag and if-range", func(t *testing.T) {
		first := request(t, httpServer, stdhttp.MethodGet, path, nil)
		etag := first.Header().Get("ETag")
		if etag == "" {
			t.Fatal("ETag is empty")
		}
		lastModified := first.Header().Get("Last-Modified")
		if lastModified == "" {
			t.Fatal("Last-Modified is empty")
		}

		notModified := request(t, httpServer, stdhttp.MethodGet, path, map[string]string{"If-None-Match": etag})
		if notModified.Code != stdhttp.StatusNotModified || notModified.Body.Len() != 0 {
			t.Fatalf("If-None-Match status/body = %d/%q, want 304 with empty body", notModified.Code, notModified.Body.String())
		}

		matching := request(t, httpServer, stdhttp.MethodGet, path, map[string]string{"Range": "bytes=2-4", "If-Range": lastModified})
		if matching.Code != stdhttp.StatusPartialContent || matching.Body.String() != "234" {
			t.Fatalf("matching If-Range status/body = %d/%q, want 206/%q", matching.Code, matching.Body.String(), "234")
		}

		stale := request(t, httpServer, stdhttp.MethodGet, path, map[string]string{"Range": "bytes=2-4", "If-Range": etag})
		if stale.Code != stdhttp.StatusOK || stale.Body.String() != content {
			t.Fatalf("weak ETag If-Range status/body = %d/%q, want 200/%q", stale.Code, stale.Body.String(), content)
		}
	})

	t.Run("method restriction", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodPost, path, nil)
		if response.Code == stdhttp.StatusOK || response.Code == stdhttp.StatusPartialContent {
			t.Fatalf("POST unexpectedly streamed video with status %d", response.Code)
		}
	})

	t.Run("missing video", func(t *testing.T) {
		response := request(t, httpServer, stdhttp.MethodGet, "/node/v1/videos/stream/missing", nil)
		if response.Code != stdhttp.StatusNotFound {
			t.Fatalf("status = %d, want %d", response.Code, stdhttp.StatusNotFound)
		}
		if strings.Contains(response.Body.String(), "Key not found") {
			t.Fatalf("response leaks repository error: %q", response.Body.String())
		}
	})
}

func newVideoHTTPServer(t *testing.T, id, name string, content []byte) (*serverHTTP, string) {
	t.Helper()

	db, err := badger.Open(badger.DefaultOptions(t.TempDir()).WithLogger(nil))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close BadgerDB: %v", err)
		}
	})

	videoPath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(videoPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	repo := data.NewVideoVideoRepository(db)
	if err := repo.Save(context.Background(), &v1.VideoMeta{Id: id, Title: name, Path: videoPath, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}

	videoService := service.NewVideoService(biz.NewVideoUsecase(repo))
	streamHandler := server.NewVideoStreamHandler(videoService)
	httpServer := server.NewHTTPServer(
		&conf.Server{Http: &conf.Server_HTTP{}},
		videoService,
		streamHandler,
		log.NewStdLogger(io.Discard),
	)
	return &serverHTTP{serve: httpServer.ServeHTTP}, "/node/v1/videos/stream/" + id
}

type serverHTTP struct {
	serve stdhttp.HandlerFunc
}

func request(t *testing.T, server *serverHTTP, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	server.serve(response, req)
	return response
}
