package collect

import (
	"encoding/json"
	"os/exec"

	"github.com/pkg/errors"
	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
)

type Aio struct {
	AioNr    string `json:"aio-nr"`
	AioMaxNr string `json:"aio-max-nr"`
}

type CollectHostAio struct {
	hostCollector *troubleshootv1beta2.Aio
}

func (c *CollectHostAio) Title() string {
	return hostCollectorTitleOrDefault(c.hostCollector.HostCollectorMeta, "Aio")
}

func (c *CollectHostAio) IsExcluded() (bool, error) {
	return isExcluded(c.hostCollector.Exclude)
}

func (c *CollectHostAio) Collect(progressChan chan<- interface{}) (map[string][]byte, error) {
	aio, err := collectAio()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read aio")
	}

	b, err := json.Marshal(aio)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal aio")
	}

	return map[string][]byte{
		"system/aio.json": b,
	}, nil
}

func collectAio() (Aio, error) {
	var aio Aio
	cmd := exec.Command("cat", "/proc/sys/fs/aio-nr")
	stdout1, err := cmd.Output()
	if err != nil {
		return aio, err
	}

	cmd = exec.Command("cat", "/proc/sys/fs/aio-max-nr")
	stdout2, err := cmd.Output()
	if err != nil {
		return aio, err
	}

	aio = Aio{AioNr: string(stdout1), AioMaxNr: string(stdout2)}

	return aio, nil

}
