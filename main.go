package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type Window struct {
	title         string
	id            string
	class         string
	monitor       string
	floating      string
	workspace     string
	workspaceName string
}

type SortBy []Window

func (a SortBy) Len() int      { return len(a) }
func (a SortBy) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortBy) Less(i, j int) bool {
	ii, _ := strconv.ParseInt(a[i].workspace, 10, 0)
	jj, _ := strconv.ParseInt(a[j].workspace, 10, 0)

	return ii < jj
}

func main() {
	if !hyprctlIsInstalled() {
		fmt.Println("Hyprctl is not installed")
		return
	}

	cmd := exec.Command("hyprctl", "clients")
	stdout, err := cmd.Output()

	if err != nil {
		log.Println(err.Error())
		return
	}

	output := string(stdout)

	windows := []Window{}

	windowStrings := strings.Split(output, "\n\n")
	for _, windowString := range windowStrings {
		if len(windowString) <= 1 {
			continue
		}

		lines := strings.Split(windowString, "\n")

		mapping := map[string]string{}
		for _, line := range lines {
			split := strings.Split(line, ": ")
			if len(split) <= 1 {
				continue
			}

			mapping[strings.TrimSpace(split[0])] = split[1]
		}

		fmt.Println()
		for k, v := range mapping {
			fmt.Println(k, v)
		}
		fmt.Println()

		wsSplit := strings.Split(mapping["workspace"], " (")

		windows = append(windows, Window{
			title:         mapping["title"],
			id:            mapping["id"],
			monitor:       mapping["monitor"],
			class:         mapping["class"],
			floating:      mapping["floating"],
			workspace:     strings.ReplaceAll(wsSplit[1], ")", ""),
			workspaceName: wsSplit[0],
		})
	}

	// for _, workspace := range workspaces {
	// 	fmt.Printf("workspace %v: \n    monitor:%v\n    windows:%v\n    title:%v\n\n", workspace.id, workspace.monitor, workspace.windows, workspace.title)
	// }

	selectedWindowWorkspace, err := getUserChoice(windows)
	if err != nil {
		log.Println(err.Error())
		return
	}

	switchCommand := exec.Command("hyprctl", "dispatch", "workspace", selectedWindowWorkspace)
	switchErr := switchCommand.Run()
	if switchErr != nil {
		log.Println(switchErr.Error())
		return
	}
}

// returns workspace id
func getUserChoice(windows SortBy) (string, error) {
	sort.Sort(windows)

	cmd := exec.Command("wofi", "--show", "dmenu", "-i")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
		return "", fmt.Errorf("could not open stdinpipe")
	}

	go func() {
		defer stdin.Close()
		for _, win := range windows {
			io.WriteString(stdin, fmt.Sprintf("%v: %v (monitor %v)\n", win.workspace, win.title, win.monitor))
		}
	}()

	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		return "", fmt.Errorf("error while executing wofi command")
	}

	id := strings.TrimSpace(strings.Split(string(out), ":")[0])

	return id, nil
}

func hyprctlIsInstalled() bool {
	_, err := exec.LookPath("hyprctl")
	return err == nil
}
