package bpf

import (
	"errors"
	"io"
)

// RingReader — l'interface minimale d'un lecteur de ring, injectable pour
// vérifier le vidage sans charger d'objet eBPF ni toucher au noyau.
type RingReader interface {
	Read() ([]byte, error)
	Close() error
}

// DrainEvents vide le ring jusqu'à io.EOF et rend le nombre d'événements lus.
// Les pertes de réservation ne se déduisent PAS d'ici : le noyau n'émet aucun
// record de perte. g est le relais hors-ring, jamais modifié par ce vidage.
func DrainEvents(r RingReader, g *Globals) (int, error) {
	n := 0
	for {
		if _, err := r.Read(); err != nil {
			if errors.Is(err, io.EOF) {
				return n, nil
			}
			return n, err
		}
		n++
	}
}
