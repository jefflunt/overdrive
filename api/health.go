package api

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ApiStartTime time.Time

func GetSchedulerStartTime() (time.Time, error) {
	data, err := os.ReadFile("scheduler_start_time")
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func RecordSchedulerStartTime() error {
	return os.WriteFile("scheduler_start_time", []byte(time.Now().Format(time.RFC3339)), 0644)
}

type ContainerInfo struct {
	Name    string
	Project string
	Type    string
}

func GetRunningContainersInfo() ([]ContainerInfo, error) {
	// Count containers running under the scheduler (non-chat containers)
	out, err := exec.Command("podman", "ps", "--format", "{{.Names}}|{{.Labels}}").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var containers []ContainerInfo
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		name := parts[0]

		labelsStr := ""
		if len(parts) > 1 {
			labelsStr = parts[1]
		}

		labels := make(map[string]string)
		for _, label := range strings.Split(labelsStr, ",") {
			kv := strings.SplitN(label, "=", 2)
			if len(kv) == 2 {
				labels[kv[0]] = kv[1]
			}
		}

		project := labels["project"]
		if project == "" {
			project = labels["com.docker.compose.project"]
		}

		cType := labels["type"]
		if cType == "" {
			cType = labels["com.docker.compose.service"]
		}

		if cType == "" {
			if strings.Contains(name, "-worker-") || strings.HasPrefix(name, "overdrive-worker-") {
				cType = "build"
			} else if strings.Contains(name, "-scheduler") {
				cType = "scheduler"
			} else if strings.HasPrefix(name, "overdrive-chat-") {
				cType = "chat"
			}
		}

		// Try to extract project from name if not in labels
		if project == "" {
			if strings.HasPrefix(name, "overdrive-chat-") {
				// overdrive-chat-${PROJECT_NAME}-${CHAT_ID}
				parts := strings.Split(name, "-")
				if len(parts) >= 3 {
					project = parts[2]
				}
			} else if strings.HasPrefix(name, "overdrive-worker-") {
				// overdrive-worker-${PROJECT_NAME}-${JOB_ID}
				parts := strings.Split(name, "-")
				if len(parts) >= 3 {
					project = parts[2]
				}
			}
		}

		displayName := name
		displayName = strings.TrimPrefix(displayName, "overdrive-chat-")
		displayName = strings.TrimPrefix(displayName, "overdrive-worker-")

		containers = append(containers, ContainerInfo{
			Name:    displayName,
			Project: project,
			Type:    cType,
		})
	}
	return containers, nil
}

func FormatUptime(duration time.Duration) string {
	d := int(duration.Hours() / 24)
	h := int(duration.Hours()) % 24
	m := int(duration.Minutes()) % 60
	return fmt.Sprintf("%dd%dh%dm", d, h, m)
}
