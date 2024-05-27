package passpersist

import (
	"fmt"
	"testing"
)

func TestNewOidEmpty(t *testing.T) {
	o := MustNewOid("1.3")

	_, err := o.Append([]int{1, 2, 3, 4})
	if err != nil {
		t.Error(err)
	}

}

func TestOidAppend(t *testing.T) {
	o := MustNewOid("1.2")

	o, _ = o.Append([]int{3, 4})
	o, _ = o.Append([]int{5, 6})
	o, _ = o.Append([]int{7, 8})

	fmt.Println(o.String())
}
