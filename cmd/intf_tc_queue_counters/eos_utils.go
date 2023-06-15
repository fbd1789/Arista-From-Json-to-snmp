package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-cmd/cmd"
)

func eosCommand(command string) []string {
	c := cmd.NewCmd("Cli", "-p", "15", "-c", command)
	c.Env = append(c.Env, "TERM=dumb")
	<-c.Start()
	return c.Status().Stdout
}

func eosCommandJson(cmd string, v any) any {
	out := eosCommand(fmt.Sprintf("%s | json", cmd))
	var buf bytes.Buffer
	for _, l := range out {
		buf.WriteString(l)
	}
	return json.Unmarshal(buf.Bytes(), v)
}

func getIfIndexeMap() map[string]int {
	indexes := make(map[string]int)
	out := eosCommand("show snmp mib walk IF-MIB::ifDescr")
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

	return indexes
}
