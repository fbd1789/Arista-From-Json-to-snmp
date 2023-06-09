package passpersist

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type SetError int

const (
	NotWriteable SetError = iota
	WrongType
	WrongValue
	WrongLength
	InconsistentValue
)

func (e SetError) String() string {
	switch e {
	case NotWriteable:
		return "not-writable"
	case WrongType:
		return "wrong-type"
	case WrongValue:
		return "wrong-value"
	case WrongLength:
		return "wrong-length"
	case InconsistentValue:
		return "inconsistent-value"
	default:
		log.Fatal().Msgf("unknown value type id: %d", e)
	}
	return ""
}

// type OID struct {
// 	Value asn1.ObjectIdentifier
// }

// func NewOid(s string) (*OID, error) {
// 	var err error
// 	subs := strings.Split(s, ".")

// 	// remove leading dot
// 	if subs[0] == "" {
// 		subs = subs[1:]
// 	}

// 	if len(subs) > 128 {
// 		return nil, errors.New("oid too long, maxuimum 128")
// 	}

// 	o := make(asn1.ObjectIdentifier, len(subs))
// 	for i, v := range subs {
// 		o[i], err = strconv.Atoi(v)
// 		if err != nil || o[i] < 0 || int64(o[i]) > math.MaxUint32 {
// 			return nil, errors.New("oid out of range.")
// 		}
// 	}

// 	return &OID{o}, nil
// }

// func (o OID) HasPrefix(oid OID) bool {
// 	if len(oid.Value) > len(o.Value) {
// 		return false
// 	}
// 	prefix := o.Value[0:len(oid.Value)]

// 	return prefix.Equal(oid.Value)
// }

// func (o OID) Equal(oid OID) bool {
// 	if len(o.Value) != len(oid.Value) {
// 		return false
// 	}
// 	for i, v := range o.Value {
// 		if v != oid.Value[i] {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (o OID) Pop(oid OID) OID {
// 	if o.HasPrefix(oid) {
// 		return OID{o.Value[len(oid.Value):]}
// 	}
// 	return OID{}
// }

// func (o OID) Prepend(oid OID) OID {
// 	return OID{append(oid.Value, o.Value...)}
// }

// func (o OID) Append(oid OID) OID {
// 	return OID{append(o.Value, oid.Value...)}
// }

// func (o OID) String() string {
// 	return o.Value.String()
// }

type VarBind struct {
	Oid       *Oid       `json:"oid"`
	ValueType string     `json:"type"`
	Value     typedValue `json:"value"`
}

func (r *VarBind) String() string {
	return fmt.Sprintf("%s, %s, %v", r.Oid, r.Value.TypeString(), r.Value)
}

func (r *VarBind) Marshal() string {

	return fmt.Sprintf("%s\n%s\n%s", r.Oid, r.Value.TypeString(), r.Value.String())
}

type typedValue struct {
	Value isTypedValue
}

func (v *typedValue) String() string {

	switch v.GetValue().(type) {
	case *StringVal:
		return v.GetStringVal()
	case *IntVal:
		return strconv.Itoa(int(v.GetIntVal()))
	case *Counter32Val:
		return strconv.FormatUint(uint64(v.GetCouter32Val()), 10)
	case *Counter64Val:
		return strconv.FormatUint(v.GetCouter64Val(), 10)
	case *GaugeVal:
		return strconv.FormatUint(uint64(v.GetGaugeVal()), 10)
	case *OctetStringVal:
		return string(v.GetOctetStringVal())
	case *IPAddrVal:
		return v.GetIPAddrVal().String()
	case *IPV6AddrVal:
		return v.GetIPV6AddrVal().String()
	case *OIDVal:
		o := v.GetOIDVal()
		return o.String()
	case *TimeTicksVal:
		return v.GetTimeTicksVal().String()
	default:
		log.Warn().Msgf("unknown value type %T", v.GetValue())
	}
	return ""
}

func (v *typedValue) TypeString() string {
	switch v.GetValue().(type) {
	case *StringVal:
		return "STRING"
	case *IntVal:
		return "INTEGER"
	case *Counter32Val:
		return "Counter32"
	case *Counter64Val:
		return "Counter64"
	case *GaugeVal:
		return "GAUGE"
	case *OctetStringVal:
		return "OCTET"
	case *IPAddrVal:
		return "IPADDRESS"
	case *IPV6AddrVal:
		return "STRING"
	case *OIDVal:
		return "OBJECTID"
	case *TimeTicksVal:
		return "TIMETICKS"
	default:
		log.Warn().Msgf("unknown value type %T", v.GetValue())
	}
	return ""
}

func (v *typedValue) GetValue() interface{} {
	if v != nil {
		return v.Value
	}
	return nil
}
func (v *typedValue) GetCouter32Val() uint32 {
	if x, ok := v.GetValue().(*Counter32Val); ok {
		return x.Value
	}
	return 0
}

func (v *typedValue) GetCouter64Val() uint64 {
	if x, ok := v.GetValue().(*Counter64Val); ok {
		return x.Value
	}
	return 0
}

func (v *typedValue) GetGaugeVal() uint32 {
	if x, ok := v.GetValue().(*GaugeVal); ok {
		return x.Value
	}
	return 0
}

func (v *typedValue) GetIntVal() int32 {
	if x, ok := v.GetValue().(*IntVal); ok {
		return x.Value
	}
	return 0
}

func (v *typedValue) GetIPAddrVal() net.IP {
	if x, ok := v.GetValue().(*IPAddrVal); ok {
		return x.Value
	}
	return net.ParseIP("0.0.0.0")
}

func (v *typedValue) GetIPV6AddrVal() net.IP {
	if x, ok := v.GetValue().(*IPV6AddrVal); ok {
		return x.Value
	}
	return net.ParseIP("::")
}

func (v *typedValue) GetOctetStringVal() []byte {
	if x, ok := v.GetValue().(*OctetStringVal); ok {
		return x.Value
	}
	return []byte{}
}

func (v *typedValue) GetOIDVal() Oid {
	if x, ok := v.GetValue().(*OIDVal); ok {
		return x.Value
	}
	return Oid{}
}

func (v *typedValue) GetStringVal() string {
	if x, ok := v.GetValue().(*StringVal); ok {
		return x.Value
	}
	return ""
}

func (v *typedValue) GetTimeTicksVal() time.Duration {
	if x, ok := v.GetValue().(*TimeTicksVal); ok {
		return x.Value
	}
	return time.Duration(0)
}

type isTypedValue interface {
	isTypedValue()
}

type Counter32Val struct {
	Value uint32
}

type Counter64Val struct {
	Value uint64
}

type GaugeVal struct {
	Value uint32
}

type IntVal struct {
	Value int32
}

type IPAddrVal struct {
	Value net.IP
}

type IPV6AddrVal struct {
	Value net.IP
}

type OctetStringVal struct {
	Value []byte
}

type OIDVal struct {
	Value Oid
}

type StringVal struct {
	Value string
}

type TimeTicksVal struct {
	Value time.Duration
}

func (*Counter32Val) isTypedValue()   {}
func (*Counter64Val) isTypedValue()   {}
func (*GaugeVal) isTypedValue()       {}
func (*IntVal) isTypedValue()         {}
func (*IPAddrVal) isTypedValue()      {}
func (*IPV6AddrVal) isTypedValue()    {}
func (*OctetStringVal) isTypedValue() {}
func (*OIDVal) isTypedValue()         {}
func (*StringVal) isTypedValue()      {}
func (*TimeTicksVal) isTypedValue()   {}
