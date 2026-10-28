package pg

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

// A dump names the roles of the server it was taken on (owners, grantees) and
// sets the parameters of the pg_dump that wrote it. Restored elsewhere, those
// roles may not exist and a newer parameter may be unknown, and with
// ON_ERROR_STOP either aborts the whole restore. RoleSafe rewrites the plain
// SQL stream so these statements cannot fail it:
//
//   - ALTER ... OWNER TO is dropped when noOwner is set (the restoring user then
//     owns the objects, as with pg_restore --no-owner);
//   - GRANT, REVOKE, ALTER DEFAULT PRIVILEGES and a kept OWNER TO run in a DO
//     block that turns a missing role or a missing right into a warning;
//   - SET of a parameter outside the long-standing pg_dump set (transaction_timeout
//     from pg_dump 17+, ...) runs the same way, so an older server skips it;
//   - the statements a pg_dump --create dump makes about its own database
//     (DROP/CREATE/ALTER DATABASE, its comment and grants) and its \connect are
//     dropped: the dump is restored into the database the restore names, as a
//     dump without --create would be, never into (or over) the source one.
//
// Only complete top-level statements on one line are touched, which is how
// pg_dump writes these; COPY data, quoted strings and dollar-quoted bodies
// pass through untouched. Every other error still stops the restore.
func RoleSafe(r io.Reader, noOwner bool) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		f := &roleFilter{noOwner: noOwner}
		err := f.run(r, pw)
		if f.dropped > 0 || f.guarded > 0 || f.database > 0 {
			log().Info("Restore adapted to this server", "owner_statements_skipped", f.dropped,
				"statements_guarded", f.guarded, "database_statements_skipped", f.database)
		}
		_ = pw.CloseWithError(err)
	}()
	return pr
}

var (
	ownerStmt     = regexp.MustCompile(`^ALTER [A-Z ]+ .+ OWNER TO .+;$`)
	privilegeStmt = regexp.MustCompile(`^(?:GRANT|REVOKE|ALTER DEFAULT PRIVILEGES) .+;$`)
	setStmt       = regexp.MustCompile(`^SET ([a-z_][a-z0-9_.]*) (?:=|TO) .+;$`)
	databaseStmt  = regexp.MustCompile(`^(?:(?:DROP|CREATE|ALTER) DATABASE|COMMENT ON DATABASE|SECURITY LABEL .*ON DATABASE|(?:GRANT|REVOKE) .* ON DATABASE) .+;$`)
	connectCmd    = regexp.MustCompile(`^\\c(?:onnect)?(?:\s|$)`)
)

// knownSettings are the parameters pg_dump has set for years, present on every
// supported server; a SET of any other one is guarded.
var knownSettings = map[string]bool{
	"statement_timeout": true, "lock_timeout": true, "idle_in_transaction_session_timeout": true,
	"client_encoding": true, "standard_conforming_strings": true, "check_function_bodies": true,
	"xmloption": true, "client_min_messages": true, "row_security": true, "search_path": true,
	"default_tablespace": true, "default_table_access_method": true, "default_with_oids": true,
	"escape_string_warning": true, "session_replication_role": true,
}

const guardTag = "$pgbkup$"

// guard wraps stmt so that a missing object (role, parameter) or a missing
// right is reported as a warning instead of failing the restore.
func guard(stmt string) string {
	return "DO " + guardTag + "BEGIN " + stmt + " EXCEPTION WHEN undefined_object OR insufficient_privilege THEN " +
		"RAISE WARNING 'pg-bkup: skipped %: %', " + quoteLiteral(stmt) + ", SQLERRM; END" + guardTag + ";"
}

func quoteLiteral(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

const lineBuf = 64 * 1024

type roleFilter struct {
	noOwner                    bool
	dropped, guarded, database int

	copying  bool   // inside the data of COPY ... FROM stdin
	quote    byte   // inside '...' or "..."
	dollar   string // inside a dollar-quoted string, its tag
	comment  int    // /* */ nesting depth
	lineCmt  bool   // inside a -- comment
	inStmt   bool   // a statement has started and not reached its ';'
	prevByte byte
}

func (f *roleFilter) run(r io.Reader, w io.Writer) error {
	br := bufio.NewReaderSize(r, lineBuf)
	partial := false // the previous chunk was a line too long for the buffer
	for {
		line, err := br.ReadSlice('\n')
		if len(line) > 0 {
			out := line
			if partial || errors.Is(err, bufio.ErrBufferFull) {
				// A long line (data, a large definition) is never one we rewrite.
				f.scanChunk(line, partial)
			} else {
				out = f.line(line)
			}
			if _, werr := w.Write(out); werr != nil {
				return werr
			}
		}
		switch {
		case errors.Is(err, bufio.ErrBufferFull):
			partial = true
		case err == nil:
			partial = false
		case errors.Is(err, io.EOF):
			return nil
		default:
			return err
		}
	}
}

// scanChunk tracks state across part of an over-long line.
func (f *roleFilter) scanChunk(b []byte, continuation bool) {
	if f.copying {
		return // a COPY row; the terminator "\." is never this long
	}
	if !continuation && f.clean() && bytes.HasPrefix(b, []byte(`\`)) {
		return // a psql meta-command
	}
	f.scan(b)
}

// line handles one complete line and returns what to write for it.
func (f *roleFilter) line(b []byte) []byte {
	text := strings.TrimRight(string(b), "\r\n")
	if f.copying {
		if text == `\.` {
			f.copying = false
		}
		return b
	}
	start := f.clean()
	if start && strings.HasPrefix(text, `\`) {
		// psql meta-command (\restrict, ...): no SQL state.
		if connectCmd.MatchString(text) {
			f.database++
			return b[len(text):]
		}
		return b
	}
	f.scan(b)
	if !start || !f.clean() {
		return b
	}
	if strings.HasPrefix(text, "COPY ") && strings.HasSuffix(text, "FROM stdin;") {
		f.copying = true
		return b
	}
	if strings.Contains(text, guardTag) {
		return b
	}
	nl := b[len(text):]
	switch {
	case databaseStmt.MatchString(text):
		f.database++
		return nl
	case ownerStmt.MatchString(text):
		if f.noOwner {
			f.dropped++
			return nl
		}
	case privilegeStmt.MatchString(text):
	case setStmt.MatchString(text):
		if knownSettings[setStmt.FindStringSubmatch(text)[1]] {
			return b
		}
	default:
		return b
	}
	f.guarded++
	return append([]byte(guard(text)), nl...)
}

// clean reports whether the stream is between statements.
func (f *roleFilter) clean() bool {
	return !f.inStmt && f.quote == 0 && f.dollar == "" && f.comment == 0 && !f.lineCmt
}

func (f *roleFilter) scan(b []byte) {
	for i := 0; i < len(b); i++ {
		c := b[i]
		prev := f.prevByte
		f.prevByte = c
		switch {
		case f.lineCmt:
			if c == '\n' {
				f.lineCmt = false
			}
		case f.dollar != "":
			if c == '$' && bytes.HasPrefix(b[i:], []byte(f.dollar)) {
				i += len(f.dollar) - 1
				f.dollar = ""
				f.prevByte = '$'
			}
		case f.quote != 0:
			if c == f.quote {
				f.quote = 0 // a doubled quote reopens on the next byte
			}
		case f.comment > 0:
			if c == '*' && i+1 < len(b) && b[i+1] == '/' {
				f.comment--
				i++
			} else if c == '/' && i+1 < len(b) && b[i+1] == '*' {
				f.comment++
				i++
			}
		default:
			switch c {
			case '-':
				if i+1 < len(b) && b[i+1] == '-' {
					f.lineCmt = true
					i++
				} else {
					f.inStmt = true
				}
			case '/':
				if i+1 < len(b) && b[i+1] == '*' {
					f.comment++
					i++
				} else {
					f.inStmt = true
				}
			case '\'', '"':
				f.quote = c
				f.inStmt = true
			case '$':
				if tag := dollarTag(b[i:]); tag != "" && !identByte(prev) {
					f.dollar = tag
					i += len(tag) - 1
					f.prevByte = '$'
				}
				f.inStmt = true
			case ';':
				f.inStmt = false
			case ' ', '\t', '\r', '\n':
			default:
				f.inStmt = true
			}
		}
	}
}

// dollarTag returns the $tag$ (or $$) opening b, or "".
func dollarTag(b []byte) string {
	for i := 1; i < len(b); i++ {
		c := b[i]
		switch {
		case c == '$':
			return string(b[:i+1])
		case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80:
		case c >= '0' && c <= '9' && i > 1:
		default:
			return ""
		}
	}
	return ""
}

func identByte(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c >= 0x80
}
