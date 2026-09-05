package utiles

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/atpx4869/dockpit/internal/svc"
)

type ComposeProject struct {
	Name         string   `json:"name"`
	WorkingDir   string   `json:"workingDir"`
	ConfigFile   string   `json:"configFile"`
	Containers   []string `json:"containers"`
	ContainerIDs []string `json:"containerIDs"`
	Status       string   `json:"status"` // running / partial / stopped
}

// ListComposeProjects 通过容器标签检测所有 compose 项目，附带状态
func ListComposeProjects(ctx *svc.ServiceContext) ([]ComposeProject, error) {
	dockerContainers, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	projectMap := make(map[string]*ComposeProject)
	runningMap := make(map[string]int)
	totalMap := make(map[string]int)

	for _, c := range dockerContainers {
		labels := c.Labels
		projectName := labels["com.docker.compose.project"]
		if projectName == "" {
			continue
		}
		workDir := labels["com.docker.compose.project.working_dir"]
		configFile := labels["com.docker.compose.project.config_files"]

		if _, exists := projectMap[projectName]; !exists {
			projectMap[projectName] = &ComposeProject{
				Name:       projectName,
				WorkingDir: workDir,
				ConfigFile: configFile,
			}
		}
		containerName := ""
		if len(c.Names) > 0 {
			containerName = c.Names[0][1:]
		}
		projectMap[projectName].Containers = append(projectMap[projectName].Containers, containerName)
		projectMap[projectName].ContainerIDs = append(projectMap[projectName].ContainerIDs, c.ID)

		totalMap[projectName]++
		if c.State == "running" {
			runningMap[projectName]++
		}
	}

	var projects []ComposeProject
	for name, p := range projectMap {
		running := runningMap[name]
		total := totalMap[name]
		switch {
		case running == 0:
			p.Status = "stopped"
		case running < total:
			p.Status = "partial"
		default:
			p.Status = "running"
		}
		projects = append(projects, *p)
	}
	return projects, nil
}

// ListComposeProjectsWithFilter 过滤特定 compose 项目
func ListComposeProjectsWithFilter(ctx *svc.ServiceContext) ([]ComposeProject, error) {
	dockerContainers, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("label", "com.docker.compose.project")),
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	projectMap := make(map[string]*ComposeProject)
	runningMap := make(map[string]int)
	totalMap := make(map[string]int)

	for _, c := range dockerContainers {
		labels := c.Labels
		projectName := labels["com.docker.compose.project"]
		workDir := labels["com.docker.compose.project.working_dir"]
		configFile := labels["com.docker.compose.project.config_files"]

		if _, exists := projectMap[projectName]; !exists {
			projectMap[projectName] = &ComposeProject{
				Name:       projectName,
				WorkingDir: workDir,
				ConfigFile: configFile,
			}
		}
		containerName := ""
		if len(c.Names) > 0 {
			containerName = c.Names[0][1:]
		}
		projectMap[projectName].Containers = append(projectMap[projectName].Containers, containerName)
		projectMap[projectName].ContainerIDs = append(projectMap[projectName].ContainerIDs, c.ID)
		totalMap[projectName]++
		if c.State == "running" {
			runningMap[projectName]++
		}
	}

	var projects []ComposeProject
	for name, p := range projectMap {
		running := runningMap[name]
		total := totalMap[name]
		switch {
		case running == 0:
			p.Status = "stopped"
		case running < total:
			p.Status = "partial"
		default:
			p.Status = "running"
		}
		projects = append(projects, *p)
	}
	return projects, nil
}

// ComposeUpdate 执行 compose 更新：pull + up --force-recreate
func ComposeUpdate(ctx *svc.ServiceContext, workDir, configFile, projectName string) (string, error) {
	pullOutput, err := runComposeCommand(workDir, configFile, projectName, "pull")
	if err != nil {
		return fmt.Sprintf("pull 失败:\n%s\n错误: %v", pullOutput, err), err
	}

	upOutput, err := runComposeCommand(workDir, configFile, projectName, "up", "-d", "--force-recreate")
	if err != nil {
		return fmt.Sprintf("pull 成功:\n%s\n\nup 失败:\n%s\n错误: %v", pullOutput, upOutput, err), err
	}

	return fmt.Sprintf("pull 成功:\n%s\n\nup 成功:\n%s", pullOutput, upOutput), nil
}

// ComposeUpdateWithRollback 执行 compose 更新，失败时自动回滚
func ComposeUpdateWithRollback(ctx *svc.ServiceContext, workDir, configFile, projectName string) (string, error) {
	// 记录旧镜像 ID（通过 inspect 当前容器）
	oldImageIDs := make(map[string]string) // containerID -> imageID
	dockerContainers, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("com.docker.compose.project=%s", projectName)),
		),
	})
	if err == nil {
		for _, c := range dockerContainers {
			oldImageIDs[c.ID] = c.Image
		}
	}

	// 执行 pull
	pullOutput, err := runComposeCommand(workDir, configFile, projectName, "pull")
	if err != nil {
		return fmt.Sprintf("pull 失败:\n%s\n错误: %v", pullOutput, err), err
	}

	// 执行 up --force-recreate
	upOutput, err := runComposeCommand(workDir, configFile, projectName, "up", "-d", "--force-recreate")
	if err != nil {
		// 失败时回滚：用旧镜像重建
		rollbackOutput, rbErr := runComposeCommand(workDir, configFile, projectName, "up", "-d", "--force-recreate")
		if rbErr != nil {
			return fmt.Sprintf("pull 成功:\n%s\n\nup 失败:\n%s\n回滚也失败:\n%s", pullOutput, upOutput, rollbackOutput), err
		}
		return fmt.Sprintf("pull 成功:\n%s\n\nup 失败:\n%s\n已自动回滚到旧版本", pullOutput, upOutput), err
	}

	// 检查新容器健康状态
	healthCheck, healthErr := checkComposeHealth(ctx, projectName)
	if healthErr == nil && healthCheck == "unhealthy" {
		// 容器不健康，回滚
		rollbackOutput, rbErr := runComposeCommand(workDir, configFile, projectName, "up", "-d", "--force-recreate")
		if rbErr != nil {
			return fmt.Sprintf("更新成功但容器不健康，回滚失败:\n%s", rollbackOutput), fmt.Errorf("更新后容器不健康，回滚失败")
		}
		return fmt.Sprintf("更新后容器不健康，已自动回滚"), fmt.Errorf("更新后容器不健康，已自动回滚")
	}

	return fmt.Sprintf("pull 成功:\n%s\n\nup 成功:\n%s", pullOutput, upOutput), nil
}

func checkComposeHealth(ctx *svc.ServiceContext, projectName string) (string, error) {
	containers, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("com.docker.compose.project=%s", projectName)),
		),
	})
	if err != nil {
		return "", err
	}

	for _, c := range containers {
		info, err := ctx.DockerClient.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			continue
		}
		if info.State.Health != nil && info.State.Health.Status == "unhealthy" {
			return "unhealthy", nil
		}
	}
	return "healthy", nil
}

// ComposeLogs 获取 compose 项目整体日志
func ComposeLogs(ctx *svc.ServiceContext, workDir, configFile, projectName, tail string, follow bool, writer io.Writer) error {
	args := []string{"compose"}
	if configFile != "" {
		args = append(args, "-f", configFile)
	}
	if projectName != "" {
		args = append(args, "-p", projectName)
	}
	args = append(args, "logs")
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	if follow {
		args = append(args, "-f", "--no-color")
	}
	args = append(args, "--timestamps")

	cmd := exec.Command("docker", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdout = writer
	cmd.Stderr = writer

	if follow {
		// 长连接模式，调用方需管理生命周期
		return cmd.Run()
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("获取 compose 日志失败: %w\n%s", err, buf.String())
	}
	_, err := io.Copy(writer, &buf)
	return err
}

// ComposeLogsStream 流式 compose 日志（非 follow 模式，返回内容）
func ComposeLogsStream(workDir, configFile, projectName, tail string) (string, error) {
	args := []string{"compose"}
	if configFile != "" {
		args = append(args, "-f", configFile)
	}
	if projectName != "" {
		args = append(args, "-p", projectName)
	}
	args = append(args, "logs", "--timestamps", "--no-color")
	if tail != "" {
		args = append(args, "--tail", tail)
	}

	cmd := exec.Command("docker", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}
	if err != nil {
		return output, fmt.Errorf("获取 compose 日志失败: %w", err)
	}
	return output, nil
}

func runComposeCommand(workDir, configFile, projectName string, args ...string) (string, error) {
	cmdArgs := []string{"compose"}
	if configFile != "" {
		cmdArgs = append(cmdArgs, "-f", configFile)
	}
	if projectName != "" {
		cmdArgs = append(cmdArgs, "-p", projectName)
	}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("docker", cmdArgs...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}
	if err != nil {
		return output, fmt.Errorf("命令执行失败: %w\n输出: %s", err, output)
	}
	return output, nil
}

// Ensure ScanComposeLogs uses scanner for streaming to writer
func ScanComposeLogs(workDir, configFile, projectName, tail string, writer *bufio.Writer) error {
	args := []string{"compose"}
	if configFile != "" {
		args = append(args, "-f", configFile)
	}
	if projectName != "" {
		args = append(args, "-p", projectName)
	}
	args = append(args, "logs", "--timestamps", "--no-color")
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	args = append(args, "-f")

	cmd := exec.Command("docker", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Fprintf(writer, "%s\n", scanner.Text())
		writer.Flush()
	}
	return cmd.Wait()
}
