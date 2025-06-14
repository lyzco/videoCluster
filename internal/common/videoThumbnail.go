package common

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
)

// GenerateThumbnail 从指定视频提取缩略图
func GenerateThumbnail(ctx context.Context, videoPath string) (string, error) {
	// 创建内存缓冲区代替文件输出
	var buf bytes.Buffer

	// 构造 ffmpeg 命令，输出到 stdout
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-i", videoPath,
		"-ss", "00:00:03", // 从3秒位置截取
		"-vframes", "1", // 只取一帧
		"-vf", "scale=320:-1", // 缩放宽度为320，高度按比例
		"-f", "image2", // 输出为图像格式
		"-c:v", "mjpeg", // 使用JPEG编码
		"-q:v", "2", // 质量参数(2-31，越小质量越高)
		"pipe:1", // 输出到标准输出
	)

	// 将标准输出重定向到内存缓冲区
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr // 错误输出到标准错误

	// 执行命令
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %v", err)
	}

	// 将二进制图像数据转为Base64
	base64Data := base64.StdEncoding.EncodeToString(buf.Bytes())

	return base64Data, nil
}
