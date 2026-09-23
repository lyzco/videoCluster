package server

import (
	"context"
	"errors"
	"fmt"
	"mime"
	stdhttp "net/http"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/biz"
)

const videoStreamPath = "/node/v1/videos/stream/{id}"

type VideoFinder interface {
	FindVideo(context.Context, string) (*v1.VideoMeta, error)
}

// VideoStreamHandler serves local video files over HTTP.
type VideoStreamHandler struct {
	videos VideoFinder
}

func NewVideoStreamHandler(videos VideoFinder) *VideoStreamHandler {
	return &VideoStreamHandler{videos: videos}
}

// Handle serves GET and HEAD requests registered by NewHTTPServer.
func (h *VideoStreamHandler) Handle(ctx http.Context) error {
	h.serve(ctx.Response(), ctx.Request(), ctx.Vars().Get("id"))
	return nil
}

func (h *VideoStreamHandler) serve(w stdhttp.ResponseWriter, r *stdhttp.Request, id string) {
	if id == "" {
		stdhttp.Error(w, "Invalid video ID", stdhttp.StatusBadRequest)
		return
	}

	video, err := h.videos.FindVideo(r.Context(), id)
	if err != nil {
		if errors.Is(err, biz.ErrVideoNotFound) {
			stdhttp.Error(w, "Video not found", stdhttp.StatusNotFound)
			return
		}
		log.Context(r.Context()).Errorf("get video info failed: %v", err)
		stdhttp.Error(w, "Unable to read video metadata", stdhttp.StatusInternalServerError)
		return
	}

	file, err := os.Open(video.Path)
	if err != nil {
		stdhttp.Error(w, "File not found", stdhttp.StatusNotFound)
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		stdhttp.Error(w, "Unable to read file stats", stdhttp.StatusInternalServerError)
		return
	}

	fileName := filepath.Base(video.Path)
	etag := fmt.Sprintf(`W/"%s-%x-%x"`, video.Id, fileStat.Size(), fileStat.ModTime().UnixNano())
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": fileName}))
	stdhttp.ServeContent(w, r, fileName, fileStat.ModTime(), file)
}
