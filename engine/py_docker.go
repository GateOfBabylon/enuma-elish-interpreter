package engine

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/engine/ops"
	. "github.com/GateOfBabylon/enuma-elish-interpreter/logger"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"github.com/docker/docker/api/types/container"
	dockerapi "github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var dockerCli *dockerclient.Client

func init() {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	dockerCli = cli
}

// executePythonTask handles Python task execution,
func executePythonTask(task *types.Task, universe *types.Universe) error {
	logger := GetLogger()
	logger.Log("Executing Python task: %q", task.Name)
	ops.ReplaceEnvsInTask(task)

	executeFunc := func() (string, error) {
		if task.PyTaskFields != nil {
			return executePythonScripts(task.PyTaskFields, universe, task.TimeStatements)
		}
		return "", executeDefaultTask(task, universe)
	}

	if task.TimeStatements != nil && task.TimeStatements.Timeout > 0 {
		to := time.Duration(task.TimeStatements.Timeout) * time.Millisecond
		logger.Log("Applying timeout: %d ms for Task: %q", task.TimeStatements.Timeout, task.Name)
		wrapped := executeFunc
		executeFunc = func() (string, error) {
			return ops.ExecuteWithTimeout(wrapped, to)
		}
	}

	retries := 0
	if task.TimeStatements != nil {
		retries = task.TimeStatements.Retry
	}
	response, err := ops.ExecuteWithRetry(executeFunc, retries)
	if err != nil {
		return fmt.Errorf("failed after %d retries for %q: %w", retries, task.Name, err)
	}

	if task.TimeStatements != nil && task.TimeStatements.Delay > 0 {
		logger.Log("Applying delay of %d ms", task.TimeStatements.Delay)
		time.Sleep(time.Duration(task.TimeStatements.Delay) * time.Millisecond)
	}

	if response != "" {
		logger.Log("Response from task %q", task.Name)
		responseList := strings.Split(response, "\n")
		for _, line := range responseList {
			logger.Log(line)
		}
	}
	response, err = ops.ExtractExportedValue(response)
	if err != nil {
		response = strings.TrimSpace(response)
	}

	if task.Export != "" && response != "" {
		return ops.ResolveExport(task.Export, response)
	}
	return nil
}

// executePythonScripts is a helper that does the actual Docker-based Python execution.
func executePythonScripts(fields *types.PyTaskFields, universe *types.Universe, ts *types.TimeStatements) (string, error) {
	logger := GetLogger()
	if universe.Secret != "" {
		decoded, err := base64.StdEncoding.DecodeString(universe.Secret)
		if err != nil {
			logger.Log("Failed to decode secret: %v", err)
			return "", fmt.Errorf("failed to decode base64 secret: %v", err)
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			logger.Log("Decoded secret doesn't appear to be 'user:pass' formatted.")
		} else {
			user, pass := credentials[0], credentials[1]
			if err := dockerLoginCLI(user, pass, universe.World); err != nil {
				logger.Log("Failed to login: %v", err)
				return "", fmt.Errorf("docker login failed: %w", err)
			}
		}
	}

	cmdArgs := buildPythonCommand(fields.ScriptPath)
	envVars := gatherTaskEnv()

	output, err := runPythonInDocker(universe.World, cmdArgs, envVars, ts)
	if err != nil {
		return "", fmt.Errorf("container execution failed: %w", err)
	}
	if err = extractAndSetEnvVars(output); err != nil {
		return "", fmt.Errorf("container execution failed: %w", err)
	}

	return output, nil

}

// extractAndSetEnvVars searches for 'export VAR=VALUE' and sets environment variables accordingly.
func extractAndSetEnvVars(output string) error {
	listOutput := strings.Split(output, "\n")
	exportRegex := regexp.MustCompile(`(?m)export\s+(\w+)\s*=\s*(.+)`)

	for _, line := range listOutput {
		matches := exportRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) != 3 {
				continue
			}
			varName := match[1]
			varValue := strings.TrimSpace(match[2])

			if err := os.Setenv(varName, varValue); err != nil {
				return fmt.Errorf("failed to set export variable %s: %w", varName, err)
			}
		}
	}
	return nil
}

// runPythonInDocker creates the container with "python <script> ..."
func runPythonInDocker(image string, cmdArgs []string, envVars []string, ts *types.TimeStatements) (string, error) {
	logger := GetLogger()
	ctx := contextWithOptionalTimeout(ts)

	pullOut, err := dockerCli.ImagePull(ctx, image, dockerapi.PullOptions{})
	if err != nil {
		logger.Log("Failed to pull image %q: %v", image, err)
		return "", fmt.Errorf("failed to pull image %q: %v", image, err)
	}
	io.Copy(io.Discard, pullOut)
	pullOut.Close()

	// Ensure script paths are inside the /app/scripts directory
	scriptCmd := append([]string{"python", "/app/" + cmdArgs[0]}, cmdArgs[1:]...)

	// Create container
	config := &container.Config{
		Image:      image,
		Cmd:        scriptCmd, // Explicitly specify full Python command
		Env:        envVars,
		WorkingDir: "/app",
	}

	resp, err := dockerCli.ContainerCreate(ctx, config, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := dockerCli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	statusCh, errCh := dockerCli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case waitErr := <-errCh:
		if waitErr != nil {
			return "", fmt.Errorf("error waiting for container: %w", waitErr)
		}
	case <-statusCh:
	}

	logsReader, err := dockerCli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logsReader.Close()

	logData, err := io.ReadAll(logsReader)
	if err != nil {
		return "", fmt.Errorf("failed to read container logs: %w", err)
	}

	if removeErr := dockerCli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); removeErr != nil {
		logger.Log("Warning: unable to remove container %s: %v\n", resp.ID, removeErr)
	}

	return string(logData), nil
}

// dockerLoginCLI is a simple demonstration using the Docker CLI to login.
// For a more robust solution, you can do a programmatic login with the Docker engine API.
func dockerLoginCLI(user, pass, image string) error {
	cmd := exec.Command("docker", "login", "-u", user, "--password-stdin")
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	// Write the password to stdin
	if _, wErr := io.WriteString(in, pass); wErr != nil {
		in.Close()
		return wErr
	}
	in.Close()
	return cmd.Wait()
}

// buildPythonCommand splits the script path by spaces.
func buildPythonCommand(scriptPath string) []string {
	return strings.Fields(scriptPath)
}

// gatherTaskEnv collects environment variables from the OS.
func gatherTaskEnv() []string {
	return os.Environ()
}

// contextWithOptionalTimeout wraps a background context with a timeout if provided.
func contextWithOptionalTimeout(ts *types.TimeStatements) context.Context {
	ctx := context.Background()
	if ts != nil && ts.Timeout > 0 {
		d := time.Duration(ts.Timeout) * time.Millisecond
		newCtx, _ := context.WithTimeout(ctx, d)
		return newCtx
	}
	return ctx
}
