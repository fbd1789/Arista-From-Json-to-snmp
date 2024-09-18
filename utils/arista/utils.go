package arista

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils"
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

	return json.Unmarshal(buf.Bytes(), &v)
}

func MustGetIfIndexeMap() map[string]int {
	m, err := GetIfIndexeMap()
	if err != nil {
		slog.Error("failed to get ifIndex map", slog.Any("error", err))
		os.Exit(1)
	}
	return m
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
		slog.Debug("adding interface index", "name", name, "idx", idx)
		indexes[name] = idx
	}

	if len(indexes) == 0 {
		return indexes, errors.New("failed to load index map")
	}
	return indexes, nil
}

// little hack to read the OID from the snmpd.conf file.  Can't think of a better way to do this yet
func GetBaseOidFromSnmpConfig() (*passpersist.Oid, error) {
	//progName := filepath.Base(os.Args[0])

	file, err := os.Open("/etc/snmp/snmpd.conf")
	//file, err := os.Open("snmpd.conf")
	if err != nil {
		return nil, errors.New("failed to read snmpd configuration file")
	}
	defer file.Close()

	re := regexp.MustCompile(`pass_persist (\.?[\.\d+]+) [\/\w+]*` + regexp.QuoteMeta(utils.ProgName()))
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m := re.FindStringSubmatch(scanner.Text())
		if len(m) > 0 {
			o := passpersist.MustNewOid(m[1])
			return &o, nil
		}

	}
	return nil, errors.New("extension not found in configutation")
}

func MustGetBaseOid() *passpersist.Oid {
	o, err := GetBaseOidFromSnmpConfig()
	if err != nil {
		slog.Error("failed to get base OID", slog.Any("error", err))
		os.Exit(1)
	}

	return o
}
