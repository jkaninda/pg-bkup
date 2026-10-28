package pg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jkaninda/bkupkit/config"
)

// Kind is the kind of a decompressed dump.
type Kind int

const (
	// KindPlain is a pg_dump plain SQL dump of one database.
	KindPlain Kind = iota
	// KindCustom is a pg_dump custom-format archive ("PGDMP" magic).
	KindCustom
	// KindCluster is a pg_dumpall dump (CREATE DATABASE, \connect).
	KindCluster
)

func (k Kind) String() string {
	switch k {
	case KindCustom:
		return FormatCustom
	case KindCluster:
		return "cluster"
	}
	return FormatPlain
}

const sniffSize = 16 * 1024

var (
	customMagic   = []byte("PGDMP")
	clusterHeader = []byte("PostgreSQL database cluster dump")
)

// Sniff peeks at the start of r and returns the dump kind and a reader that
// still yields every byte. file (the artifact name) helps recognise a
// pg_dumpall dump whose header was stripped.
func Sniff(r io.Reader, file string) (Kind, io.Reader, error) {
	br := bufio.NewReaderSize(r, sniffSize)
	head, err := br.Peek(sniffSize)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
		return KindPlain, nil, fmt.Errorf("read backup: %w", err)
	}
	if len(bytes.TrimSpace(head)) == 0 {
		return KindPlain, nil, errors.New("the backup is empty")
	}
	switch {
	case bytes.HasPrefix(head, customMagic):
		return KindCustom, br, nil
	case bytes.Contains(head, clusterHeader), strings.HasPrefix(file, AllInOneName+"_"):
		return KindCluster, br, nil
	}
	return KindPlain, br, nil
}

// RestoreCommand returns the client and arguments for restoring a dump of
// kind into db. The data is read from stdin.
func RestoreCommand(v *config.Values, kind Kind, db string) (string, []string) {
	single := v.Bool(SingleTransaction)
	if kind == KindCustom {
		args := []string{"-w", "-d", ConnInfo(db), "--exit-on-error"}
		if single {
			args = append(args, "--single-transaction")
		}
		if v.Bool(RestoreNoOwner) {
			args = append(args, "--no-owner")
		}
		if v.Bool(Clean) {
			args = append(args, "--clean", "--if-exists")
		}
		return "pg_restore", args
	}
	args := []string{"-X", "-q", "-w", "-d", ConnInfo(db)}
	stop := v.String(OnErrorStop)
	if stop == "true" || (stop != "false" && kind != KindCluster) {
		args = append(args, "-v", "ON_ERROR_STOP=1")
	}
	// A pg_dumpall dump runs CREATE DATABASE and \connect: it cannot run in
	// one transaction.
	if single && kind != KindCluster {
		args = append(args, "--single-transaction")
	}
	return "psql", args
}
