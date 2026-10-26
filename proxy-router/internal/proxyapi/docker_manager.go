package proxyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/moby/go-archive"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/jsonmessage"
)

// buildProgressWidth is the bar width the old jsonmessage.JSONProgress.String used without a terminal
const buildProgressWidth = 200

// DockerManager is a wrapper for the Docker SDK
type DockerManager struct {
	client *client.Client
	log    lib.ILogger
}

// ContainerInfo represents details about a Docker container
type ContainerInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"`
	State       string            `json:"state"`
	CreatedAt   string            `json:"createdAt"`
	Ports       []PortMapping     `json:"ports"`
	Labels      map[string]string `json:"labels"`
	NetworkMode string            `json:"networkMode"`
}

// PortMapping represents a port mapping for a container
type PortMapping struct {
	HostIP        string `json:"hostIp"`
	HostPort      string `json:"hostPort"`
	ContainerPort string `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

// BuildProgress represents the progress of a Docker image build
type BuildProgress struct {
	Status       string  `json:"status"`
	Stream       string  `json:"stream,omitempty"`
	Progress     string  `json:"progress,omitempty"`
	ID           string  `json:"id,omitempty"`
	Current      int64   `json:"current,omitempty"`
	Total        int64   `json:"total,omitempty"`
	Percentage   float64 `json:"percentage,omitempty"`
	Error        string  `json:"error,omitempty"`
	TimeUpdated  int64   `json:"timeUpdated"`
	ErrorDetails string  `json:"errorDetails,omitempty"`
}

// BuildProgressCallback is a function that reports build progress
type BuildProgressCallback func(progress BuildProgress) error

// NewDockerManager creates a new Docker manager
func NewDockerManager(log lib.ILogger) *DockerManager {
	fmt.Println("Creating Docker client with env", os.Getenv("DOCKER_HOST"))
	// The client negotiates the API version with the daemon by default
	cli, err := client.New(client.FromEnv)
	if err != nil {
		log.Error("Error creating Docker client:", err)
		return &DockerManager{client: nil, log: log}
	}
	return &DockerManager{client: cli, log: log}
}

// IsClientReady checks if the Docker client is ready
func (d *DockerManager) IsClientReady() error {
	if d.client == nil {
		return fmt.Errorf("Docker client is not ready")
	}
	return nil
}

// BuildImage builds a Docker image from a Dockerfile in the specified context path
func (d *DockerManager) BuildImage(ctx context.Context, contextPath, dockerfile, imageName, imageTag string, buildArgs map[string]string, progressCallback BuildProgressCallback) (string, error) {
	if err := d.IsClientReady(); err != nil {
		return "", err
	}

	// Create tag
	tag := imageName
	if imageTag != "" {
		tag = fmt.Sprintf("%s:%s", imageName, imageTag)
	}

	// Prepare the build context
	buildContextTar, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContextTar.Close()

	// Prepare build options
	buildOptions := client.ImageBuildOptions{
		Context:    buildContextTar,
		Dockerfile: dockerfile,
		Tags:       []string{tag},
		Remove:     true,
	}

	// Add build args if provided
	if len(buildArgs) > 0 {
		options := make(map[string]*string)
		for k, v := range buildArgs {
			value := v
			options[k] = &value
		}
		buildOptions.BuildArgs = options
	}

	// Build the image
	resp, err := d.client.ImageBuild(ctx, buildContextTar, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	decoder := json.NewDecoder(resp.Body)
	for {
		var msg jsonstream.Message
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode build response: %w", err)
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Report progress
		if progressCallback != nil {
			fmt.Println("Progress message", msg)
			progressStr := ""
			if msg.Progress != nil {
				progressStr = jsonmessage.RenderTUIProgress(*msg.Progress, buildProgressWidth)
			}
			progress := BuildProgress{
				Status:      msg.Status,
				Stream:      msg.Stream,
				Progress:    progressStr,
				ID:          msg.ID,
				TimeUpdated: time.Now().UnixMilli(),
			}

			if msg.Error != nil {
				progress.Error = msg.Error.Message
				progress.ErrorDetails = msg.Error.Message
			}

			if progressStr != "" && msg.ID != "" {
				// Extract progress numbers if available
				if progressStr != "" {
					// Usually, progress is reported as: "300B/1.2MB"
					progress.Status = "building"
				}
			}

			err := progressCallback(progress)
			if err != nil {
				return "", fmt.Errorf("progress callback failed: %w", err)
			}
		}
	}

	// Verify the image was built
	_, err = d.client.ImageInspect(ctx, tag)
	if err != nil {
		return "", fmt.Errorf("built image not found, build may have failed: %w", err)
	}

	return tag, nil
}

// StartContainer starts a Docker container with the specified options
func (d *DockerManager) StartContainer(ctx context.Context, imageName string, containerName string, env []string, ports map[string]string, volumes map[string]string, networkMode string) (string, error) {
	if err := d.IsClientReady(); err != nil {
		return "", err
	}

	// Create port bindings
	portBindings := network.PortMap{}
	exposedPorts := network.PortSet{}
	for containerPort, hostPort := range ports {
		port, err := network.ParsePort(containerPort)
		if err != nil {
			return "", fmt.Errorf("invalid container port %q: %w", containerPort, err)
		}
		// "8080" and "8080/tcp" parse to the same port; map order would pick the binding
		if _, ok := portBindings[port]; ok {
			return "", fmt.Errorf("container port %q is listed more than once", port.String())
		}
		exposedPorts[port] = struct{}{}
		portBindings[port] = []network.PortBinding{
			{
				HostIP:   netip.IPv4Unspecified(),
				HostPort: hostPort,
			},
		}
	}

	// Create volume bindings
	var binds []string
	for hostPath, containerPath := range volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	// Create container configuration
	config := &container.Config{
		Image:        imageName,
		ExposedPorts: exposedPorts,
		Env:          env,
	}

	// Create host configuration
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
	}

	// Set network mode if specified
	if networkMode != "" {
		hostConfig.NetworkMode = container.NetworkMode(networkMode)
	}

	// Create container
	resp, err := d.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     config,
		HostConfig: hostConfig,
		Name:       containerName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if _, err := d.client.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

// StopContainer stops a Docker container
func (d *DockerManager) StopContainer(ctx context.Context, containerID string, timeout int) error {
	if err := d.IsClientReady(); err != nil {
		return err
	}

	// Set reasonable default timeout
	if timeout <= 0 {
		timeout = 10
	}

	_, err := d.client.ContainerStop(ctx, containerID, client.ContainerStopOptions{
		Timeout: &timeout,
	})
	return err
}

// RemoveContainer removes a Docker container
func (d *DockerManager) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	if err := d.IsClientReady(); err != nil {
		return err
	}

	options := client.ContainerRemoveOptions{
		Force: force,
	}

	_, err := d.client.ContainerRemove(ctx, containerID, options)
	return err
}

// GetContainerInfo retrieves information about a container
func (d *DockerManager) GetContainerInfo(ctx context.Context, containerID string) (*ContainerInfo, error) {
	if err := d.IsClientReady(); err != nil {
		return nil, err
	}

	result, err := d.client.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	resp := result.Container

	// Extract port mappings
	ports := []PortMapping{}
	for containerPort, bindings := range resp.NetworkSettings.Ports {
		for _, binding := range bindings {
			ports = append(ports, PortMapping{
				HostIP:        addrString(binding.HostIP),
				HostPort:      binding.HostPort,
				ContainerPort: containerPort.Port(),
				Protocol:      string(containerPort.Proto()),
			})
		}
	}

	info := &ContainerInfo{
		ID:          resp.ID,
		Name:        resp.Name,
		Image:       resp.Config.Image,
		Status:      string(resp.State.Status),
		State:       string(resp.State.Status),
		CreatedAt:   resp.Created,
		Ports:       ports,
		Labels:      resp.Config.Labels,
		NetworkMode: string(resp.HostConfig.NetworkMode),
	}

	return info, nil
}

// ListContainers lists all containers matching the specified filters
func (d *DockerManager) ListContainers(ctx context.Context, all bool, filterLabels map[string]string) ([]ContainerInfo, error) {
	if err := d.IsClientReady(); err != nil {
		return nil, err
	}

	// Create filters
	filterArgs := make(client.Filters)
	for k, v := range filterLabels {
		filterArgs.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	// List containers
	list, err := d.client.ContainerList(ctx, client.ContainerListOptions{
		All:     all,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Process containers
	result := make([]ContainerInfo, 0, len(list.Items))
	for _, c := range list.Items {
		// Extract port mappings
		ports := []PortMapping{}
		for _, port := range c.Ports {
			ports = append(ports, PortMapping{
				HostIP:        addrString(port.IP),
				HostPort:      fmt.Sprintf("%d", port.PublicPort),
				ContainerPort: fmt.Sprintf("%d", port.PrivatePort),
				Protocol:      port.Type,
			})
		}

		// Get first name without leading slash
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		// Create container info
		info := ContainerInfo{
			ID:        c.ID,
			Name:      name,
			Image:     c.Image,
			Status:    c.Status,
			State:     string(c.State),
			CreatedAt: time.Unix(c.Created, 0).Format(time.RFC3339),
			Ports:     ports,
			Labels:    c.Labels,
		}

		result = append(result, info)
	}

	return result, nil
}

// GetContainerLogs retrieves logs from a container
func (d *DockerManager) GetContainerLogs(ctx context.Context, containerID string, tail int, follow bool) (io.ReadCloser, error) {
	if err := d.IsClientReady(); err != nil {
		return nil, err
	}

	options := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
	}

	if tail > 0 {
		options.Tail = fmt.Sprintf("%d", tail)
	}

	return d.client.ContainerLogs(ctx, containerID, options)
}

// GetDockerVersion gets the Docker server version
func (d *DockerManager) GetDockerVersion(ctx context.Context) (string, error) {
	if err := d.IsClientReady(); err != nil {
		return "", err
	}

	versionInfo, err := d.client.ServerVersion(ctx, client.ServerVersionOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get Docker version: %w", err)
	}

	return versionInfo.Version, nil
}

// PruneImages removes unused Docker images
func (d *DockerManager) PruneImages(ctx context.Context) (int64, error) {
	if err := d.IsClientReady(); err != nil {
		return 0, err
	}

	result, err := d.client.ImagePrune(ctx, client.ImagePruneOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune images: %w", err)
	}

	return int64(result.Report.SpaceReclaimed), nil
}

// PruneContainers removes stopped containers
func (d *DockerManager) PruneContainers(ctx context.Context) (int64, error) {
	if err := d.IsClientReady(); err != nil {
		return 0, err
	}

	result, err := d.client.ContainerPrune(ctx, client.ContainerPruneOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune containers: %w", err)
	}

	return int64(result.Report.SpaceReclaimed), nil
}

// addrString returns "" for an unset address, as the string fields of the old API types did
func addrString(addr netip.Addr) string {
	if !addr.IsValid() {
		return ""
	}
	return addr.String()
}
