package passpersist

import (
	"context"
	"testing"
	"time"
)

// func TestReadStdin(t *testing.T) {

// 	ctx := context.Background()
// 	pp := NewPassPersist(5)

// 	go pp.Run(ctx, func(*PassPersist) {})

// 	<-ctx.Done()
// }

func TestPopOID(t *testing.T) {

	base := Config.BaseOid
	oid, _ := OIDFromString("99.99")
	oid = append(base, oid...)
	suffix, _ := OIDFromString("99.99")
	got := oid.Pop(base)

	if got == nil {
		t.Errorf("failed to cut OID: %s", oid.String())
	} else if !got.Equal(suffix) {
		t.Errorf("OID.Cut = %s; want %s", got.String(), suffix.String())
	}

}

func TestCallback(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	Config.Refresh = time.Second
	p := NewPassPersist(&Config)
	go p.update(ctx, func(pp *PassPersist) {
		pp.AddString([]int{255, 0}, "Hello")
		pp.AddInt([]int{255, 1}, 42)
		pp.AddString([]int{255, 2}, "!")
	})
	p.Dump()
	select {
	case <-time.After(time.Second * 2):
		p.Dump()
	case <-ctx.Done():
		return
	}
}

func TestAddString(t *testing.T) {

}
