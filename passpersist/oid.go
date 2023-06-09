package passpersist

// taken from: https://github.com/k-sone/snmpgo/blob/master/variables.go

import (
	"bytes"
	"encoding/asn1"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/geoffgarside/ber"
)

type InvalidOidErr struct {
	Value   string
	Message string
}

func (e *InvalidOidErr) Error() string {
	return fmt.Sprintf("Invalid OID '%s': %s", e.Value, e.Message)
}

type Oid struct {
	Value asn1.ObjectIdentifier
}

func (v *Oid) String() string {
	return v.Value.String()
}

func (v *Oid) Type() string {
	return "Oid"
}

func (v *Oid) Marshal() ([]byte, error) {
	return asn1.Marshal(v.Value)
}

func (v *Oid) Unmarshal(b []byte) (rest []byte, err error) {
	var i asn1.ObjectIdentifier
	rest, err = ber.Unmarshal(b, &i)
	if err == nil {
		v.Value = i
	}
	return
}

// Returns true if this OID contains the specified OID
func (v *Oid) Contains(o *Oid) bool {
	if o == nil || len(v.Value) < len(o.Value) {
		return false
	}
	for i := 0; i < len(o.Value); i++ {
		if v.Value[i] != o.Value[i] {
			return false
		}
	}
	return true
}

// Returns 0 this OID is equal to the specified OID,
// -1 this OID is lexicographically less than the specified OID,
// 1 this OID is lexicographically greater than the specified OID
func (v *Oid) Compare(o *Oid) int {
	if o != nil {
		vl := len(v.Value)
		ol := len(o.Value)
		for i := 0; i < vl; i++ {
			if ol <= i || v.Value[i] > o.Value[i] {
				return 1
			} else if v.Value[i] < o.Value[i] {
				return -1
			}
		}
		if ol == vl {
			return 0
		}
	}
	return -1
}

// Returns true if this OID is same the specified OID
func (v *Oid) Equal(o *Oid) bool {
	if o == nil {
		return false
	}
	return v.Value.Equal(o.Value)
}

// Returns Oid with additional sub-ids
func (v *Oid) Append(subs []int) (*Oid, error) {
	buf := bytes.NewBufferString(v.String())
	for _, i := range subs {
		buf.WriteString(".")
		buf.WriteString(strconv.Itoa(i))
	}
	return NewOid(buf.String())
}

func NewOid(s string) (oid *Oid, err error) {
	subids := strings.Split(s, ".")

	// if there is a leadong dot, the first element will be empty
	if subids[0] == "" {
		subids = subids[1:]
	}

	// RFC2578 Section 3.5
	if len(subids) > 128 {
		return nil, &InvalidOidErr{s, "The sub-identifiers in an OID is up to 128"}
	}

	o := make(asn1.ObjectIdentifier, len(subids))
	for i, v := range subids {
		o[i], err = strconv.Atoi(v)
		if err != nil || o[i] < 0 || int64(o[i]) > math.MaxUint32 {
			return nil, &InvalidOidErr{s, fmt.Sprintf("The sub-identifiers is range %d..%d", 0, int64(math.MaxUint32))}
		}
	}

	if len(o) > 0 && o[0] > 2 {
		return nil, &InvalidOidErr{s, "The first sub-identifier is range 0..2"}
	}

	// ISO/IEC 8825 Section 8.19.4
	if len(o) < 2 {
		return nil, &InvalidOidErr{s, "The first and second sub-identifier is required"}
	}

	if o[0] < 2 && o[1] >= 40 {
		return nil, &InvalidOidErr{s, "The second sub-identifier is range 0..39"}
	}

	return &Oid{o}, nil
}

// MustNewOid is like NewOid but panics if argument cannot be parsed
func MustNewOid(s string) *Oid {
	if oid, err := NewOid(s); err != nil {
		panic(`snmpgo.MustNewOid: ` + err.Error())
	} else {
		return oid
	}
}

type Oids []*Oid

// Sort a Oid list
func (o Oids) Sort() Oids {
	c := make(Oids, len(o))
	copy(c, o)
	sort.Sort(sortableOids{c})
	return c
}

func (o Oids) uniq(comp func(a, b *Oid) bool) Oids {
	var before *Oid
	c := make(Oids, 0, len(o))
	for _, oid := range o {
		if !comp(before, oid) {
			before = oid
			c = append(c, oid)
		}
	}
	return c
}

// Filter out adjacent OID list
func (o Oids) Uniq() Oids {
	return o.uniq(func(a, b *Oid) bool {
		if b == nil {
			return a == nil
		} else {
			return b.Equal(a)
		}
	})
}

// Filter out adjacent OID list with the same prefix
func (o Oids) UniqBase() Oids {
	return o.uniq(func(a, b *Oid) bool {
		if b == nil {
			return a == nil
		} else {
			return b.Contains(a)
		}
	})
}

type sortableOids struct {
	Oids
}

func (o sortableOids) Len() int {
	return len(o.Oids)
}

func (o sortableOids) Swap(i, j int) {
	o.Oids[i], o.Oids[j] = o.Oids[j], o.Oids[i]
}

func (o sortableOids) Less(i, j int) bool {
	return o.Oids[i] != nil && o.Oids[i].Compare(o.Oids[j]) < 1
}

func NewOids(s []string) (oids Oids, err error) {
	for _, l := range s {
		o, e := NewOid(l)
		if e != nil {
			return nil, e
		}
		oids = append(oids, o)
	}
	return
}
