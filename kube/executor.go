package kube

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/c-bata/kube-prompt/internal/debug"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

func (c *Completer) Executor(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	} else if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	// 对于线上环境屏蔽一些操作
	if blockDangerousProdCommand(c, s) {
		return
	}

	cmd := exec.Command("/bin/sh", "-c", "kubectl "+s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Got error: %s\n", err.Error())
	}

	if changeContext(s) {
		GlobalCompleter.ReloadContext()
		ReloadContextInfo(GlobalCompleter)
	}

	return
}

func ExecuteAndGetResult(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		debug.Log("you need to pass the something arguments")
		return ""
	}

	out := &bytes.Buffer{}
	cmd := exec.Command("/bin/sh", "-c", "kubectl "+s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		debug.Log(err.Error())
		return ""
	}
	r := string(out.Bytes())
	return r
}

func changeContext(command string) bool {
	args := strings.Split(command, " ")
	return len(args) > 2 && args[0] == "config" && args[1] == "use-context"
}

func blockDangerousProdCommand(c *Completer, commandStr string) (isDangerousProdCommand bool) {
	args := strings.Split(commandStr, " ")
	command := args[0]

	if strings.Contains(c.context, "test") || strings.Contains(c.context, "beta") {
		return
	}

	if strings.Contains("edit;scale;delete;apply;create;replace;rollout;rolling-update;run;", command) {
		color.Red("您现在正在非test/beta环境执行危险指令，请确认后在rancher中执行!!")
		// 解析cluster url
		// https://rancher.domain.com/k8s/clusters/<c-id>  ==>  https://rancher.domain.com/dashboard/c/<c-id>/explorer
		dashboardUrl := c.clusterUrl
		pattern := `(http[s]://.*\..*\..*)/k8s/clusters/(.*)`
		re, err := regexp.Compile(pattern)
		if err == nil {
			submatch := re.FindAllSubmatch([]byte(c.clusterUrl), 1)
			if len(submatch) > 0 {
				domain := submatch[0][1]
				cId := submatch[0][2]
				dashboardUrl = fmt.Sprintf("%s/dashboard/c/%s/explorer", domain, cId)
			}
		}

		renderResult([][]string{
			{"command", "kubectl " + commandStr},
			{"rancher", dashboardUrl},
		})
		isDangerousProdCommand = true
	}
	return
}

func renderResult(contents [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetRowSeparator("-")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(contents)
	table.SetAutoWrapText(true)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.Render()
}
