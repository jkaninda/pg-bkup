package pg

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func filtered(t *testing.T, in string, noOwner bool) string {
	t.Helper()
	rc := RoleSafe(iotest.OneByteReader(strings.NewReader(in)), noOwner)
	defer func() { _ = rc.Close() }()
	out, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestRoleSafe(t *testing.T) {
	in := `\restrict abc
SET statement_timeout = 0;
SET transaction_timeout = 0;
SELECT pg_catalog.set_config('search_path', '', false);
-- Name: users; Type: TABLE; Schema: public; Owner: u_83c8ac41
CREATE TABLE public.users (
    id integer NOT NULL,
    name text DEFAULT 'GRANT x TO y;'::text
);
ALTER TABLE public.users OWNER TO u_83c8ac41;
COPY public.users (id, name) FROM stdin;
1	GRANT SELECT ON t TO r;
2	ALTER TABLE t OWNER TO r;
\.
CREATE FUNCTION public.f() RETURNS void
    LANGUAGE plpgsql
    AS $_$
BEGIN
GRANT SELECT ON public.users TO ro;
END
$_$;
ALTER FUNCTION public.f() OWNER TO u_83c8ac41;
GRANT SELECT ON TABLE public.users TO ro_83c8ac41;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE u_83c8ac41 IN SCHEMA public GRANT SELECT ON TABLES TO ro_83c8ac41;
\unrestrict abc
`
	out := filtered(t, in, true)
	for _, keep := range []string{
		"\\restrict abc\n", "SET statement_timeout = 0;\n", "name text DEFAULT 'GRANT x TO y;'::text\n",
		"1\tGRANT SELECT ON t TO r;\n2\tALTER TABLE t OWNER TO r;\n\\.\n",
		"BEGIN\nGRANT SELECT ON public.users TO ro;\nEND\n", "\\unrestrict abc\n",
	} {
		if !strings.Contains(out, keep) {
			t.Errorf("lost %q:\n%s", keep, out)
		}
	}
	if strings.Contains(out, "OWNER TO u_83c8ac41") {
		t.Errorf("owner statements kept:\n%s", out)
	}
	for _, g := range []string{
		"SET transaction_timeout = 0;", "GRANT SELECT ON TABLE public.users TO ro_83c8ac41;",
		"REVOKE ALL ON SCHEMA public FROM PUBLIC;",
		"ALTER DEFAULT PRIVILEGES FOR ROLE u_83c8ac41 IN SCHEMA public GRANT SELECT ON TABLES TO ro_83c8ac41;",
	} {
		if !strings.Contains(out, guard(g)+"\n") {
			t.Errorf("not guarded: %q:\n%s", g, out)
		}
	}
	if strings.Count(out, "DO $pgbkup$") != 4 {
		t.Errorf("want 4 guarded statements:\n%s", out)
	}

	kept := filtered(t, "ALTER TABLE public.users OWNER TO u_1;\n", false)
	if kept != guard("ALTER TABLE public.users OWNER TO u_1;")+"\n" {
		t.Errorf("RESTORE_NO_OWNER=false: %q", kept)
	}
}

func TestRoleSafeLongLines(t *testing.T) {
	long := strings.Repeat("x", 3*lineBuf)
	in := "INSERT INTO t VALUES ('" + long + "');\nGRANT SELECT ON t TO r;\nSELECT '" + long + "\nGRANT SELECT ON t TO r;\n';\n"
	out := filtered(t, in, true)
	want := "INSERT INTO t VALUES ('" + long + "');\n" + guard("GRANT SELECT ON t TO r;") + "\nSELECT '" + long + "\nGRANT SELECT ON t TO r;\n';\n"
	if out != want {
		t.Fatalf("long lines mangled (len %d, want %d)", len(out), len(want))
	}
}

func TestGuardQuotes(t *testing.T) {
	g := guard(`GRANT SELECT ON TABLE public."it's" TO r;`)
	if !strings.Contains(g, `'GRANT SELECT ON TABLE public."it''s" TO r;'`) {
		t.Fatal(g)
	}
}

func TestRoleSafeCreateDump(t *testing.T) {
	in := `\restrict k
DROP DATABASE IF EXISTS posta;
CREATE DATABASE posta WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8';
ALTER DATABASE posta OWNER TO u_83c8ac41;
COMMENT ON DATABASE posta IS 'mail';
GRANT CONNECT ON DATABASE posta TO ro_83c8ac41;
ALTER DATABASE posta SET search_path TO public;
\unrestrict k
\connect posta
\c posta
\restrict k
CREATE TABLE public.t (i integer);
\unrestrict k
`
	want := `\restrict k






\unrestrict k


\restrict k
CREATE TABLE public.t (i integer);
\unrestrict k
`
	if out := filtered(t, in, true); out != want {
		t.Fatalf("got:\n%s", out)
	}
	// A \copy or \connection-like command that is not \connect is kept.
	if out := filtered(t, "\\copyright\n", true); out != "\\copyright\n" {
		t.Fatalf("%q", out)
	}
}
