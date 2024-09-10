package arista

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/go-cmd/cmd"
	"github.com/rs/zerolog/log"
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

// little hack to read the OID from the snmpd.conf file.  Can't think of a better way to do this yet
func GetBaseOid() (passpersist.Oid, error) {
	progName := filepath.Base(os.Args[0])

	file, err := os.Open("/etc/snmp/snmpd.conf")
	//file, err := os.Open("snmpd.conf")
	if err != nil {
		return passpersist.Oid{}, errors.New("failed to read snmpd configuration file")
	}
	defer file.Close()

	re := regexp.MustCompile(`pass_persist (\.?[\.\d+]+) [\/\w+]*` + regexp.QuoteMeta(progName))
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m := re.FindStringSubmatch(scanner.Text())
		if len(m) > 0 {
			return passpersist.MustNewOid(m[1]), nil
		}

	}
	return passpersist.Oid{}, errors.New("extension not found in configutation")
}

func MustGetBaseOid() passpersist.Oid {
	o, err := GetBaseOid()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get base OID")
	}

	return o
}
