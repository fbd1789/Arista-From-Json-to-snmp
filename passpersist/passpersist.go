package passpersist

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"runtime"
	"time"

	"os"

	"log/slog"

	"golang.org/x/sys/unix"
)

type SetError int

const (
	NotWriteable SetError = iota
	// WrongType
	// WrongValue
	// WrongLength
	// InconsistentValue
)

func (e SetError) String() string {
	switch e {
	case NotWriteable:
		return "not-writable"
	// case WrongType:
	// 	return "wrong-type"
	// case WrongValue:
	// 	return "wrong-value"
	// case WrongLength:
	// 	return "wrong-length"
	// case InconsistentValue:
	// 	return "inconsistent-value"
	default:
		slog.Error("unknown value type id", slog.Any("error", e))
	}
	return "unknown-error"
}

type PassPersist struct {
	ctx    context.Context
	cache  *Cache
	config Config
}

func NewPassPersist(ctx context.Context, opts ...ConfigFunc) *PassPersist {
	c := NewConfigWithDefaults(ctx)

	for _, fn := range opts {
		fn(&c)
	}

	return &PassPersist{
		ctx:    ctx,
		cache:  NewCache(),
		config: c,
	}
}

func (p *PassPersist) get(oid OID) *VarBind {
	slog.Debug("getting oid", "oid", oid.String())
	return p.cache.Get(oid)
}

func (p *PassPersist) getNext(oid OID) *VarBind {
	return p.cache.GetNext(oid)
}

func (p *PassPersist) AddEntry(subs []int, value typedValue) error {
	oid, err := p.config.BaseOID.Append(subs)
	if err != nil {
		return err
	}

	slog.Debug("adding entry", "type", value.TypeString(), "oid", oid.String(), "value", value.String())

	err = p.cache.Set(&VarBind{
		OID:       oid,
		ValueType: value.TypeString(),
		Value:     value,
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *PassPersist) AddString(subIds []int, value string) error {
	return p.AddEntry(subIds, typedValue{&StringVal{value}})
}

func (p *PassPersist) AddInt(subIds []int, value int32) error {
	return p.AddEntry(subIds, typedValue{&IntVal{value}})
}

func (p *PassPersist) AddOID(subIds []int, value OID) error {
	return p.AddEntry(subIds, typedValue{&OIDVal{value}})
}

func (p *PassPersist) AddOctetString(subIds []int, value []byte) error {
	return p.AddEntry(subIds, typedValue{&OctetStringVal{value}})
}

func (p *PassPersist) AddIP(subIds []int, value netip.Addr) error {
	return p.AddEntry(subIds, typedValue{&IPAddrVal{value}})
}

func (p *PassPersist) AddIPV6(subIds []int, value netip.Addr) error {
	return p.AddEntry(subIds, typedValue{&IPV6AddrVal{value}})
}

func (p *PassPersist) AddCounter32(subIds []int, value uint32) error {
	return p.AddEntry(subIds, typedValue{&Counter32Val{value}})
}

func (p *PassPersist) AddCounter64(subIds []int, value uint64) error {
	return p.AddEntry(subIds, typedValue{&Counter64Val{value}})
}

func (p *PassPersist) AddGauge(subIds []int, value uint32) error {
	return p.AddEntry(subIds, typedValue{&GaugeVal{value}})
}

func (p *PassPersist) AddTimeTicks(subIds []int, value time.Duration) error {
	return p.AddEntry(subIds, typedValue{&TimeTicksVal{value}})
}

func (p *PassPersist) Dump() {
	out := make(map[string]interface{})

	out["base-oid"] = p.config.BaseOID
	out["refresh"] = p.config.RefreshInterval

	j, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(j))

	p.cache.Dump()
}

func setPrio(prio int) error {
	var err error

	switch runtime.GOOS {
	case "linux", "bsd", "freebsd", "netbsd", "openbsd":
		err = unix.Setpriority(unix.PRIO_PROCESS, 0, prio)
	}

	if err != nil {
		return err
	}

	return nil
}

func (p *PassPersist) update(ctx context.Context, callback func(*PassPersist)) {

	err := setPrio(15)
	if err != nil {
		slog.Warn("failed to set priority")
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			timer := time.NewTimer(p.config.RefreshInterval)

			callback(p)
			p.cache.Commit()

			<-timer.C
		}
	}
}

func (p *PassPersist) Run(f func(*PassPersist)) {
	input := make(chan string)
	done := make(chan bool)

	go p.update(p.ctx, f)
	go watchStdin(p.ctx, input, done)

	for {
		select {
		case line := <-input:
			switch line {
			case "PING":
				fmt.Println("PONG")
			case "getnext":
				inp := <-input
				slog.Debug("validating", "input", inp)
				if oid, ok := p.convertAndValidateOID(inp); ok {
					slog.Debug("getNext", "oid", oid.String())
					v := p.getNext(oid)
					if v != nil {
						fmt.Println(v.Marshal())
					} else {
						fmt.Println("NONE")
					}
				} else {
					slog.Warn("failed to validate input", "input", inp)
					fmt.Println("NONE")
				}

			case "get":
				inp := <-input
				if oid, ok := p.convertAndValidateOID(inp); ok {
					slog.Debug("get", "oid", oid.String())
					v := p.get(oid)
					if v != nil {
						fmt.Println(v.Marshal())
					} else {
						fmt.Println("NONE")
					}
				} else {
					fmt.Println("NONE")
				}
			case "set":
				fmt.Println(NotWriteable)
			case "DUMP", "D":
				p.Dump()
			case "DUMPCACHE", "DC":
				p.cache.Dump()
			case "DUMPINDEX", "DI":
				p.cache.DumpIndex()
			default:
				fmt.Println("NONE")
			}
		case <-done:
			return
		case <-p.ctx.Done():
			return
		}
	}
}

func watchStdin(ctx context.Context, input chan<- string, done chan<- bool) {

	scanner := bufio.NewScanner(os.Stdin)

	defer func() {
		done <- true
	}()

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			slog.Debug("got user input", "input", line)
			input <- line
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			slog.Error("scanner encountered an error", slog.Any("error", err.Error()))
		}
	}
}

func (p *PassPersist) convertAndValidateOID(oid string) (OID, bool) {
	o, err := NewOID(oid)

	if err != nil {
		slog.Warn("failed to load oid", "oid", oid)
		return OID{}, false
	}

	if !o.Contains(p.config.BaseOID) {
		return o, false
	}

	return o, true
}
