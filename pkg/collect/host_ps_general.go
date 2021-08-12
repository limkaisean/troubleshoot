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

type PsGeneral struct {
	Pid     string `json:"pid"`
	Uname   string `json:"uname"`
	Ppid    string `json:"ppid"`
	Pgid    string `json:"pgid"`
	Sid     string `json:"sid"`
	Sz      string `json:"sz"`
	Rssize  string `json:"rssize"`
	Vsize   string `json:"vsize"`
	Psr     string `json:"psr"`
	C       string `json:"c"`
	Bsdtime string `json:"bsdtime"`
	Nlwp    string `json:"nlwp"`
	Lstart  string `json:"lstart"`
	Etimes  string `json:"etimes"`
	State   string `json:"state"`
	Tname   string `json:"tname"`
	Args    string `json:"args"`
}

type CollectHostPsGeneral struct {
	hostCollector *troubleshootv1beta2.PsGeneral
}

func (c *CollectHostPsGeneral) Title() string {
	return hostCollectorTitleOrDefault(c.hostCollector.HostCollectorMeta, "Ps General")
}

func (c *CollectHostPsGeneral) IsExcluded() (bool, error) {
	return isExcluded(c.hostCollector.Exclude)
}

func (c *CollectHostPsGeneral) Collect(progressChan chan<- interface{}) (map[string][]byte, error) {
	data, err := collectPsGeneral()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read ps general")
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ps general")
	}

	return map[string][]byte{
		"ps-general": b,
	}, nil
}

func collectPsGeneral() ([]PsGeneral, error) {
	cmd := exec.Command("ps", "-ewwo", "pid,uname,ppid,pgid,sid,sz,rssize,vsize,psr,c,bsdtime,nlwp,lstart,etimes,state,tname,args")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(stdout)
	scanner := bufio.NewScanner(buf)
	scanner.Scan()

	var data []PsGeneral
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) < 21 {
			continue
		}

		data = append(data, PsGeneral{
			Pid:     words[0],
			Uname:   words[1],
			Ppid:    words[2],
			Pgid:    words[3],
			Sid:     words[4],
			Sz:      words[5],
			Rssize:  words[6],
			Vsize:   words[7],
			Psr:     words[8],
			C:       words[9],
			Bsdtime: words[10],
			Nlwp:    words[11],
			Lstart:  strings.Join(words[12:17], " "),
			Etimes:  words[17],
			State:   words[18],
			Tname:   words[19],
			Args:    strings.Join(words[20:], " "),
		})
	}

	return data, nil
}
