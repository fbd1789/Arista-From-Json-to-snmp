package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arista-northwest/go-passpersist/passpersist"
)

var oid passpersist.OID = passpersist.MustNewOID(`1.3.6.1.4.1.8072.2.255.226`)

func writeSnmpdConfig(path string) error {
	config := strings.Join([]string{`pass_persist`, oid.String(), ProgPath()}, " ") // `pass_persist ` + oid + ` ` + ProgPath()
	return os.WriteFile(path, []byte(config), 0600)
}

func Test_GetBaseOIDFromSnmpdConfig(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, `snmpd.conf`)
	if err := writeSnmpdConfig(path); err != nil {
		t.Error("failed generate config", err)
	}

	got, err := getBaseOIDFromSNMPdConfig(path)
	if err != nil {
		t.Error(err)
	}

	if !got.Equal(oid) {
		t.Errorf("base oid does not match: %s != %s", got, oid)
	}
}
