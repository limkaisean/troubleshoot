package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Mount struct {
	Device     string   `json:"device"`
	MountPoint string   `json:"mount_point"`
	FsType     string   `json:"fs_type"`
	Modes      []string `json:"modes"`
}

func main() {
	cmd := exec.Command("cat", "/proc/mounts")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
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

	b, _ := json.Marshal(mounts)
	fmt.Println(string(b))
}
