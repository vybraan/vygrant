package storage

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"golang.org/x/oauth2"
)

type EventStore struct {
	inner TokenStore
	cmd   string
}

func NewEventStore(inner TokenStore, cmd string) *EventStore {
	return &EventStore{inner: inner, cmd: cmd}
}

func (e *EventStore) Inner() TokenStore {
	return e.inner
}

func (e *EventStore) Set(account string, token *oauth2.Token) error {
	if err := e.inner.Set(account, token); err != nil {
		return err
	}
	e.trigger(account, "set")
	e.trigger(account, "change")
	return nil
}

func (e *EventStore) Get(account string) (*oauth2.Token, error) {
	return e.inner.Get(account)
}

func (e *EventStore) Delete(account string) error {
	if err := e.inner.Delete(account); err != nil {
		return err
	}
	e.trigger(account, "delete")
	e.trigger(account, "change")
	return nil
}

func (e *EventStore) ListAccounts() []string {
	return e.inner.ListAccounts()
}

func (e *EventStore) Dump() ([]byte, error) {
	if dumper, ok := e.inner.(TokenDumper); ok {
		return dumper.Dump()
	}
	return nil, os.ErrInvalid
}

func (e *EventStore) Restore(data []byte) error {
	if dumper, ok := e.inner.(TokenDumper); ok {
		if err := dumper.Restore(data); err != nil {
			return err
		}
		e.trigger("", "restore")
		e.trigger("", "change")
		return nil
	}
	return os.ErrInvalid
}

func (e *EventStore) trigger(account, event string) {
	if e.cmd == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", e.cmd)
	cmd.Env = append(os.Environ(),
		"VYGRANT_ACCOUNT="+account,
		"VYGRANT_EVENT="+event,
	)
	if err := cmd.Run(); err != nil {
		log.Printf("token_event_cmd failed: %v", err)
	}
}
