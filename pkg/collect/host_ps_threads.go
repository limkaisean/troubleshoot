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

type PsThreads struct {
	Uid   string `json:"uid"`
	Pid   string `json:"pid"`
	Ppid  string `json:"ppid"`
	Pgid  string `json:"pgid"`
	Sid   string `json:"sid"`
	Lwp   string `json:"lwp"`
	C     string `json:"c"`
	Nlwp  string `json:"nlwp"`
	Sz    string `json:"sz"`
	Rss   string `json:"rss"`
	Psr   string `json:"psr"`
	Stime string `json:"stime"`
	Tty   string `json:"tty"`
	Time  string `json:"time"`
	Cmd   string `json:"cmd"`
}

type CollectHostPsThreads struct {
	hostCollector *troubleshootv1beta2.PsThreads
}

func (c *CollectHostPsThreads) Title() string {
	return hostCollectorTitleOrDefault(c.hostCollector.HostCollectorMeta, "Ps Threads")
}

func (c *CollectHostPsThreads) IsExcluded() (bool, error) {
	return isExcluded(c.hostCollector.Exclude)
}

func (c *CollectHostPsThreads) Collect(progressChan chan<- interface{}) (map[string][]byte, error) {
	data, err := collectPsThreads()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read ps threads")
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ps threads")
	}

	return map[string][]byte{
		"ps-threads": b,
	}, nil
}

func collectPsThreads() ([]PsThreads, error) {
	cmd := exec.Command("ps", "-ejFwwL")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data []PsThreads
	buf := bytes.NewBuffer(stdout)
	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) < 15 {
			continue
		}
		data = append(data, PsThreads{
			Uid:   words[0],
			Pid:   words[1],
			Ppid:  words[2],
			Pgid:  words[3],
			Sid:   words[4],
			Lwp:   words[5],
			C:     words[6],
			Nlwp:  words[7],
			Sz:    words[8],
			Rss:   words[9],
			Psr:   words[10],
			Stime: words[11],
			Tty:   words[12],
			Time:  words[13],
			Cmd:   strings.Join(words[14:], " "),
		})
	}

	return data, nil
}
