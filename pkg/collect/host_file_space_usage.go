package collect

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
)

type FileSpaceUsage struct {
	Usage string `json:"usage"`
	Path  string `json:"path"`
}

type CollectHostFileSpaceUsage struct {
	hostCollector *troubleshootv1beta2.FileSpaceUsage
}

func (c *CollectHostFileSpaceUsage) Title() string {
	return hostCollectorTitleOrDefault(c.hostCollector.HostCollectorMeta, "File Space Usage")
}

func (c *CollectHostFileSpaceUsage) IsExcluded() (bool, error) {
	return isExcluded(c.hostCollector.Exclude)
}

func (c *CollectHostFileSpaceUsage) Collect(progressChan chan<- interface{}) (map[string][]byte, error) {
	data, err := collectFileSpaceUsage(c.hostCollector.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file space usage")
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal file space usage")
	}

	return map[string][]byte{
		"system/file_space_usage.json": b,
	}, nil
}

func collectFileSpaceUsage(input string) (map[string][]FileSpaceUsage, error) {
	fileSpaceUsage := make(map[string][]FileSpaceUsage)

	path := "."
	if len(input) > 0 {
		path = input
	}

	cmd := exec.Command("du", "-b", "-a", path)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data []FileSpaceUsage
	buf := bytes.NewBuffer(stdout)
	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) != 2 {
			continue
		}
		data = append(data, FileSpaceUsage{
			Usage: words[0],
			Path:  words[1],
		})
	}

	fileSpaceUsage[path] = data

	return fileSpaceUsage, nil
}
