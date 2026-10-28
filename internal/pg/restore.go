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

// PsqlArgs returns the psql arguments for replaying a dump of kind into db,
// read from stdin. A custom-format archive reaches psql as the SQL script
// pg_restore makes of it (ScriptArgs), so both formats restore the same way.
func PsqlArgs(v *config.Values, kind Kind, db string) []string {
	args := []string{"-X", "-q", "-w", "-d", ConnInfo(db)}
	stop := v.String(OnErrorStop)
	if stop == "true" || (stop != "false" && kind != KindCluster) {
		args = append(args, "-v", "ON_ERROR_STOP=1")
	}
	// A pg_dumpall dump runs CREATE DATABASE and \connect: it cannot run in
	// one transaction.
	if v.Bool(SingleTransaction) && kind != KindCluster {
		args = append(args, "--single-transaction")
	}
	return args
}

// ScriptArgs returns the pg_restore arguments that turn a custom-format
// archive read from stdin into a SQL script on stdout.
func ScriptArgs(v *config.Values) []string {
	args := []string{"-f", "-"}
	if v.Bool(RestoreNoOwner) {
		args = append(args, "--no-owner")
	}
	if v.Bool(Clean) {
		args = append(args, "--clean", "--if-exists")
	}
	return args
}
