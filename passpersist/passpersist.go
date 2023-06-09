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

func (p *PassPersist) AddString(subs []int, value string) error {
	return p.AddEntry(subs, typedValue{Value: &StringVal{Value: value}})
}

func (p *PassPersist) AddInt(subs []int, value int) error {
	return p.AddEntry(subs, typedValue{Value: &IntVal{Value: value}})
}

func (p *PassPersist) AddOID(subs []int, value Oid) error {
	return nil
}

func (p *PassPersist) AddOctetString(subs []int, value []byte) error {
	return nil
}

func (p *PassPersist) AddIP(subs []int, ip net.IP) error {
	return nil
}

func (p *PassPersist) AddCounter32(subs []int, value int32) error {
	return p.AddEntry(subs, typedValue{Value: &Counter32Val{Value: value}})
}

func (p *PassPersist) AddCounter64(subs []int, value int64) error {
	return p.AddEntry(subs, typedValue{Value: &Counter64Val{Value: value}})
}

func (p *PassPersist) AddGauge(subs []int, value int) error {
	return p.AddEntry(subs, typedValue{Value: &GaugeVal{Value: value}})
}

func (p *PassPersist) AddTimeTicks(subs []int, value time.Duration) error {
	return nil
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
			log.Debug().Msg("ctx done")
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
	return o, true
}
