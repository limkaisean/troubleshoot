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

type DiskFree struct {
	Filesystem string `json:"filesystem"`
	Type       string `json:"type"`
	Blocks     string `json:"1k_blocks"`
	Used       string `json:"used"`
	Available  string `json:"available"`
	UsePercent string `json:"use_percent"`
	MountedOn  string `json:"mounted_on"`
}

type CollectHostDiskFree struct {
	hostCollector *troubleshootv1beta2.DiskFree
}

func (c *CollectHostDiskFree) Title() string {
	return hostCollectorTitleOrDefault(c.hostCollector.HostCollectorMeta, "Disk Free")
}

func (c *CollectHostDiskFree) IsExcluded() (bool, error) {
	return isExcluded(c.hostCollector.Exclude)
}

func (c *CollectHostDiskFree) Collect(progressChan chan<- interface{}) (map[string][]byte, error) {
	data, err := collectDiskFree()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read disk free")
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal disk free")
	}

	return map[string][]byte{
		"disk_free": b,
	}, nil
}

func collectDiskFree() ([]DiskFree, error) {
	cmd := exec.Command("df", "--print-type")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data []DiskFree
	buf := bytes.NewBuffer(stdout)
	scanner := bufio.NewScanner(buf)
	scanner.Scan()

	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) < 7 {
			continue
		}
		data = append(data, DiskFree{
			Filesystem: words[0],
			Type:       words[1],
			Blocks:     words[2],
			Used:       words[3],
			Available:  words[4],
			UsePercent: words[5],
			MountedOn:  words[6],
		})
	}

	return data, nil
}
