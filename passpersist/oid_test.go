package passpersist

import (
	"fmt"
	"testing"
)

func TestNewOIDEmpty(t *testing.T) {
	o := MustNewOID("1.3")

	_, err := o.Append([]int{1, 2, 3, 4})
	if err != nil {
		t.Error(err)
	}

}

func TestOIDAppend(t *testing.T) {
	o := MustNewOID("1.2")

	o, _ = o.Append([]int{3, 4})
	o, _ = o.Append([]int{5, 6})
	o, _ = o.Append([]int{7, 8})

	fmt.Println(o.String())
}
