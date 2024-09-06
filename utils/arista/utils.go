package arista

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-cmd/cmd"
)

func EosCommand(command string) ([]string, error) {
	c := cmd.NewCmd("Cli", "-p15", "-c", command)
	c.Env = append(c.Env, "TERM=dumb")
	<-c.Start()

	stderr := c.Status().Stderr
	if len(stderr) > 0 {
		return []string{}, fmt.Errorf("%s", strings.Join(stderr, "\n"))
	}

	return c.Status().Stdout, nil
}

func EosCommandJson(command string, v any) error {
	out, err := EosCommand(fmt.Sprintf("%s | json", command))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	for _, l := range out {
		buf.WriteString(l)
	}

	return json.Unmarshal(buf.Bytes(), v)
}

func GetIfIndexeMap() (map[string]int, error) {
	indexes := make(map[string]int)
	out, err := EosCommand("show snmp mib walk IF-MIB::ifDescr")
	if err != nil {
		return nil, err
	}

	for _, l := range out {
		re := regexp.MustCompile(`IF-MIB::ifDescr\[(\d+)\] = STRING: ([^$]+)`)
		t := re.FindStringSubmatch(string(l))
		if t == nil {
			continue
		}
		idx, _ := strconv.Atoi(t[1])
		name := t[2]
		indexes[name] = idx
	}

	return indexes, nil
}
