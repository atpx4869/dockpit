package utiles

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/atpx4869/dockpit/internal/svc"
)

// ExecInContainer 在容器内执行命令并返回输出
func ExecInContainer(ctx *svc.ServiceContext, id string, cmd []string) (string, error) {
	config := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execIDResp, err := ctx.DockerClient.ContainerExecCreate(context.Background(), id, config)
	if err != nil {
		return "", fmt.Errorf("创建 exec 失败: %w", err)
	}

	attachResp, err := ctx.DockerClient.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("附加 exec 失败: %w", err)
	}
	defer attachResp.Close()

	var outBuf, errBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader)
	if err != nil {
		return "", fmt.Errorf("读取 exec 输出失败: %w", err)
	}

	output := outBuf.String()
	if errBuf.Len() > 0 {
		output += "\n" + errBuf.String()
	}
	return output, nil
}

// ExecInContainerWithStream 流式 exec 输出（WebSocket 用）
func ExecInContainerWithStream(ctx *svc.ServiceContext, id string, cmd []string, writer io.Writer) error {
	config := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execIDResp, err := ctx.DockerClient.ContainerExecCreate(context.Background(), id, config)
	if err != nil {
		return fmt.Errorf("创建 exec 失败: %w", err)
	}

	attachResp, err := ctx.DockerClient.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("附加 exec 失败: %w", err)
	}
	defer attachResp.Close()

	_, err = stdcopy.StdCopy(writer, writer, attachResp.Reader)
	return err
}

// BatchContainerOperation 批量操作容器（停止/重启/删除）
func BatchContainerOperation(ctx *svc.ServiceContext, ids []string, action string) (map[string]string, error) {
	results := make(map[string]string)

	for _, id := range ids {
		var err error
		switch action {
		case "stop":
			timeout := 10
			err = ctx.DockerClient.ContainerStop(context.Background(), id, container.StopOptions{Timeout: &timeout})
		case "restart":
			timeout := 10
			err = ctx.DockerClient.ContainerRestart(context.Background(), id, container.StopOptions{Timeout: &timeout})
		case "start":
			err = ctx.DockerClient.ContainerStart(context.Background(), id, container.StartOptions{})
		case "remove":
			err = ctx.DockerClient.ContainerRemove(context.Background(), id, container.RemoveOptions{Force: true})
		default:
			results[id] = "未知操作"
			continue
		}

		if err != nil {
			results[id] = fmt.Sprintf("失败: %v", err)
		} else {
			results[id] = "成功"
		}
	}
	return results, nil
}
