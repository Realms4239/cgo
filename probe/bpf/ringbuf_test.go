package bpf

import (
	"errors"
	"io"
	"testing"
)

// fakeRingReader — lecteur injectable : une file d'événements et d'erreurs,
// sans charger d'objet eBPF ni toucher au noyau.
type fakeRingReader struct {
	items []ringItem
	i     int
}

type ringItem struct {
	data []byte
	err  error
}

func (f *fakeRingReader) Read() ([]byte, error) {
	if f.i >= len(f.items) {
		return nil, io.EOF
	}
	it := f.items[f.i]
	f.i++
	return it.data, it.err
}

func (f *fakeRingReader) Close() error { return nil }

// Oracle positif — deux événements puis io.EOF : le vidage compte 2 et ne
// touche pas au compteur de réservation, qui vit hors du ring.
func TestDrainEventsCountsUntilEOF(t *testing.T) {
	g := &Globals{ReserveFailures: 7}
	r := &fakeRingReader{items: []ringItem{
		{data: []byte("even-1")},
		{data: []byte("even-2")},
	}}
	n, err := DrainEvents(r, g)
	if err != nil {
		t.Fatalf("DrainEvents: %v", err)
	}
	if n != 2 {
		t.Fatalf("n=%d, attendu 2", n)
	}
	if g.ReserveFailures != 7 {
		t.Fatalf("le vidage a modifié le compteur hors-ring: %d", g.ReserveFailures)
	}
}

// Oracle hostile — une erreur au deuxième read : elle remonte et le nombre
// d'événements déjà lus est rendu, sans les avaler.
func TestDrainEventsPropagatesErrorWithPartialCount(t *testing.T) {
	boom := errors.New("ring reader cassé")
	r := &fakeRingReader{items: []ringItem{
		{data: []byte("even-1")},
		{err: boom},
	}}
	n, err := DrainEvents(r, &Globals{})
	if !errors.Is(err, boom) {
		t.Fatalf("err=%v, attendu %v", err, boom)
	}
	if n != 1 {
		t.Fatalf("n=%d, attendu 1 événement déjà lu", n)
	}
}
