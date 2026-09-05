package utiles

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/atpx4869/dockpit/internal/svc"
)

type ContainerStats struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CPUPercent   float64 `json:"cpuPercent"`
	MemUsage     string  `json:"memUsage"`
	MemLimit     string  `json:"memLimit"`
	MemPercent   float64 `json:"memPercent"`
	NetRx        string  `json:"netRx"`
	NetTx        string  `json:"netTx"`
	BlockRead    string  `json:"blockRead"`
	BlockWrite   string  `json:"blockWrite"`
	PIDs         int     `json:"pids"`
}

// GetContainerStats 获取单个容器的资源统计
func GetContainerStats(ctx *svc.ServiceContext, id string) (*ContainerStats, error) {
	stats, err := ctx.DockerClient.ContainerStatsOneShot(context.Background(), id)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var v types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return nil, err
	}

	// CPU 使用率计算
	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)
	var cpuPercent float64
	if systemDelta > 0 && cpuDelta >= 0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	// 内存
	memUsage := v.MemoryStats.Usage
	if v.MemoryStats.Stats != nil {
		if cache, ok := v.MemoryStats.Stats["cache"]; ok {
			memUsage -= cache
		}
	}
	memLimit := v.MemoryStats.Limit
	var memPercent float64
	if memLimit > 0 {
		memPercent = float64(memUsage) / float64(memLimit) * 100.0
	}

	// 网络
	var netRx, netTx uint64
	for _, net := range v.Networks {
		netRx += net.RxBytes
		netTx += net.TxBytes
	}

	// Block I/O
	var blockRead, blockWrite uint64
	for _, bio := range v.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(bio.Op) {
		case "read":
			blockRead += bio.Value
		case "write":
			blockWrite += bio.Value
		}
	}

	name := v.Name
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}

	return &ContainerStats{
		ID:         v.ID,
		Name:       name,
		CPUPercent: cpuPercent,
		MemUsage:   formatBytes(memUsage),
		MemLimit:   formatBytes(memLimit),
		MemPercent: memPercent,
		NetRx:      formatBytes(netRx),
		NetTx:      formatBytes(netTx),
		BlockRead:  formatBytes(blockRead),
		BlockWrite: formatBytes(blockWrite),
		PIDs:       v.PidsStats.Current,
	}, nil
}

// GetAllContainerStats 获取所有容器的资源统计
func GetAllContainerStats(ctx *svc.ServiceContext) ([]ContainerStats, error) {
	containers, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var statsList []ContainerStats
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0][1:]
		}
		if c.State != "running" {
			statsList = append(statsList, ContainerStats{
				ID:   c.ID,
				Name: name,
			})
			continue
		}
		st, err := GetContainerStats(ctx, c.ID)
		if err != nil {
			statsList = append(statsList, ContainerStats{ID: c.ID, Name: name})
			continue
		}
		statsList = append(statsList, *st)
	}
	return statsList, nil
}

// StreamContainerLogs 实时日志流（WebSocket 用）
func StreamContainerLogs(ctx *svc.ServiceContext, id string, tail string, writer io.Writer) error {
	reader, err := ctx.DockerClient.ContainerLogs(context.Background(), id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Timestamps: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// 去掉 8 字节 Docker 日志流头
		if len(line) > 8 {
			line = line[8:]
		}
		fmt.Fprintf(writer, "%s\n", line)
	}
	return scanner.Err()
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
