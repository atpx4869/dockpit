package utiles

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/docker/docker/api/types/container"
	"github.com/atpx4869/dockpit/internal/svc"
)

// ComposeProject 代表一个 docker compose 项目
type ComposeProject struct {
	Name         string   `json:"name"`
	WorkingDir   string   `json:"workingDir"`
	ConfigFile   string   `json:"configFile"`
	Containers   []string `json:"containers"`
	ContainerIDs []string `json:"containerIDs"`
}

// ListComposeProjects 通过容器标签检测所有 compose 项目
func ListComposeProjects(ctx *svc.ServiceContext) ([]ComposeProject, error) {
	dockerContainers, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	projectMap := make(map[string]*ComposeProject)

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
	}

	var projects []ComposeProject
	for _, p := range projectMap {
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
