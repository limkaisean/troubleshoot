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

type Mount struct {
	Device     string   `json:"device"`
	MountPoint string   `json:"mount_point"`
	FsType     string   `json:"fs_type"`
	Modes      []string `json:"modes"`
}

type CollectHostProcMounts struct {
	hostCollector *troubleshootv1beta2.ProcMounts
}

func (c *CollectHostProcMounts) Title() string {
	return hostCollectorTitleOrDefault(c.hostCollector.HostCollectorMeta, "Proc Mounts")
}

func (c *CollectHostProcMounts) IsExcluded() (bool, error) {
	return isExcluded(c.hostCollector.Exclude)
}

func (c *CollectHostProcMounts) Collect(progressChan chan<- interface{}) (map[string][]byte, error) {
	mounts, err := collectInfo()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read proc mounts")
	}

	b, err := json.Marshal(mounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal proc mounts")
	}

	return map[string][]byte{
		"system/proc_mounts.json": b,
	}, nil
}

func collectInfo() (map[string][]Mount, error) {
	data := make(map[string][]Mount)

	cmd := exec.Command("cat", "/proc/mounts")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var mounts []Mount
	buf := bytes.NewBuffer(stdout)
	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		if len(words) != 6 {
			continue
		}
		mounts = append(mounts, Mount{
			Device:     words[0],
			MountPoint: words[1],
			FsType:     words[2],
			Modes:      strings.Split(words[3], ","),
		})
	}

	data["mounts"] = mounts

	return data, nil
}
