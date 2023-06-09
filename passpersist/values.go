package passpersist

import (
	"net"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

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
