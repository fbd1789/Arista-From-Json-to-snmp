package passpersist

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"runtime"
	"time"

	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

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

type PassPersist struct {
	baseOid *Oid
	refresh time.Duration
	cache   *Cache
}

func NewPassPersist(config *ConfigT) *PassPersist {
	return &PassPersist{
		baseOid: config.BaseOid,
		refresh: config.Refresh,
		cache:   NewCache(),
	}
}

func (p *PassPersist) get(oid *Oid) *VarBind {
	log.Debug().Msgf("getting oid: %s", oid.String())
	return p.cache.Get(oid)
}

func (p *PassPersist) getNext(oid *Oid) *VarBind {
	return p.cache.GetNext(oid)
}

func (p *PassPersist) AddEntry(subs []int, value typedValue) error {
	oid, err := p.baseOid.Append(subs)
	if err != nil {
		return err
	}

	log.Debug().Msgf("adding %s: %s, %s", value.TypeString(), oid, value)

	err = p.cache.Set(&VarBind{
		Oid:       oid,
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

func (p *PassPersist) AddOID(subIds []int, value Oid) error {
	return p.AddEntry(subIds, typedValue{&OIDVal{value}})
}

func (p *PassPersist) AddOctetString(subIds []int, value []byte) error {
	return p.AddEntry(subIds, typedValue{&OctetStringVal{value}})
}

func (p *PassPersist) AddIP(subIds []int, value net.IP) error {
	return p.AddEntry(subIds, typedValue{&IPAddrVal{value}})
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

	out["base-oid"] = p.baseOid
	out["refresh"] = p.refresh

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
		log.Warn().Msgf("failed to set priority")
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			timer := time.NewTimer(p.refresh)

			callback(p)
			p.cache.Commit()

			<-timer.C
		}
	}
}

func (p *PassPersist) Run(ctx context.Context, f func(*PassPersist)) {
	input := make(chan string)
	done := make(chan bool)

	go p.update(ctx, f)
	go watchStdin(ctx, input, done)

	for {
		select {
		case line := <-input:
			switch line {
			case "PING":
				fmt.Println("PONG")
			case "getnext":
				inp := <-input
				if oid, ok := p.convertAndValidateOid(inp); ok {
					v := p.getNext(oid)
					if v != nil {
						fmt.Println(v.Marshal())
					} else {
						fmt.Println("NONE")
					}
				} else {
					fmt.Println("NONE")
				}

			case "get":
				inp := <-input
				if oid, ok := p.convertAndValidateOid(inp); ok {
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
				// not-writable, wrong-type, wrong-length, wrong-value or inconsistent-value
				fmt.Println("not-writable")
			case "DUMP", "D":
				p.Dump()
			case "DUMPCACHE", "DC":
				p.cache.Dump()
			default:
				fmt.Println("NONE")
			}
		case <-done:
			return
		case <-ctx.Done():
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
			log.Debug().Msgf("Got user input: %s", line)
			input <- line
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			log.Error().Msg(err.Error())
		}
	}
}

func (p *PassPersist) convertAndValidateOid(oid string) (*Oid, bool) {
	o, err := NewOid(oid)

	if err != nil {
		return nil, false
	}

	if !o.Contains(p.baseOid) {
		return o, false
	}

	return o, true
}
