package utiles

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/atpx4869/dockpit/internal/svc"
)

type ContainerLogResult struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// GetContainerLogs 获取容器最后 N 行日志
func GetContainerLogs(ctx *svc.ServiceContext, id string, tail int) (*ContainerLogResult, error) {
	tailStr := strconv.Itoa(tail)
	if tail <= 0 {
		tailStr = "100"
	}

	reader, err := ctx.DockerClient.ContainerLogs(context.Background(), id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tailStr,
		Timestamps: true,
	})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}

	return &ContainerLogResult{
		Content:   buf.String(),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
