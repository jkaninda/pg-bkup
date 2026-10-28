package pg

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/jackc/pgx/v5"
	"github.com/jkaninda/bkupkit/cli"
	"github.com/jkaninda/bkupkit/config"
	"github.com/jkaninda/bkupkit/engine"
	"github.com/spf13/pflag"
)

func registry() *config.Registry {
	r := config.NewRegistry()
	config.RegisterCommon(r)
	New().Settings(r)
	return r
}

func values(t *testing.T, env map[string]string) *config.Values {
	t.Helper()
	v, err := registry().Resolve(config.ResolveOptions{Env: config.MapEnv(env)})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestParseURL(t *testing.T) {
	cases := []struct {
		in   string
		want URL
	}{
		{"postgres://u:p%40ss%3Aw%2Frd@db.example:5433/shop?sslmode=require",
			URL{Host: "db.example", Port: "5433", User: "u", Password: "p@ss:w/rd", Database: "shop", SSLMode: "require"}},
		{"postgresql://u@localhost/app", URL{Host: "localhost", User: "u", Database: "app"}},
		{"jdbc:postgresql://h:5432/db?user=bob&password=s%26cr%3Dt",
			URL{Host: "h", Port: "5432", User: "bob", Password: "s&cr=t", Database: "db"}},
		{"jdbc:postgresql://h/db?user=bob&password=x&ssl=true", URL{Host: "h", User: "bob", Password: "x", Database: "db", SSLMode: "require"}},
		{"postgres://u:p@h/db?sslmode=verify-full&sslrootcert=/ca.pem",
			URL{Host: "h", User: "u", Password: "p", Database: "db", SSLMode: "verify-full", SSLRootCert: "/ca.pem"}},
	}
	for _, c := range cases {
		got, err := ParseURL(c.in)
		if err != nil {
			t.Fatalf("%s: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("%s:\n got %+v\nwant %+v", c.in, got, c.want)
		}
	}
	for _, bad := range []string{"mysql://h/db", "postgres:///db", "postgres://u:p@ss@h:x/db"} {
		if _, err := ParseURL(bad); err == nil {
			t.Errorf("%s: expected an error", bad)
		}
	}
}

func TestSourceConnPrecedence(t *testing.T) {
	v := values(t, map[string]string{
		"DB_URL": "postgres://url-user:url%20pw@url-host:6543/urldb?sslmode=require", "DB_HOST": "ignored",
	})
	c, err := SourceConn(v)
	if err != nil {
		t.Fatal(err)
	}
	want := Conn{Host: "url-host", Port: "6543", User: "url-user", Password: "url pw", Database: "urldb", AuthDB: "postgres", SSLMode: "require"}
	if c != want {
		t.Fatalf("got %+v", c)
	}
	// An explicit database and sslmode win over the URL.
	v = values(t, map[string]string{"DB_URL": "jdbc:postgresql://h:5432/urldb?user=u&password=p", "DB_NAME": "other", "DB_SSL_MODE": "disable", "DB_AUTH_DATABASE": "maint"})
	c, _ = SourceConn(v)
	if c.Database != "other" || c.SSLMode != "disable" || c.AuthDB != "maint" || c.User != "u" || c.Password != "p" {
		t.Fatalf("got %+v", c)
	}
	// Plain variables, DB_USER alias.
	v = values(t, map[string]string{"DB_HOST": "pg", "DB_USER": "me", "DB_PASSWORD": "pw", "DB_NAME": "d"})
	c, _ = SourceConn(v)
	if c.Host != "pg" || c.Port != "5432" || c.User != "me" || c.Database != "d" || c.SSLMode != "" {
		t.Fatalf("got %+v", c)
	}
}

func TestSSLModeValidated(t *testing.T) {
	if _, err := registry().Resolve(config.ResolveOptions{Env: config.MapEnv{"DB_SSL_MODE": "sometimes"}}); err == nil {
		t.Fatal("expected an invalid sslmode error")
	}
	v := values(t, map[string]string{"DB_SSL_MODE": "Verify-Full"})
	if v.String(config.DBSSLMode) != "verify-full" {
		t.Fatal(v.String(config.DBSSLMode))
	}
	v = values(t, map[string]string{"DB_URL": "postgres://u@h/d?sslmode=bogus"})
	if err := New().Configure(t.Context(), v); err == nil {
		t.Fatal("expected Configure to reject the URL sslmode")
	}
}

func TestTargetConn(t *testing.T) {
	v := values(t, map[string]string{
		"TARGET_DB_HOST": "t", "TARGET_DB_PORT": "5433", "TARGET_DB_USERNAME": "tu", "TARGET_DB_PASSWORD": "tp",
		"TARGET_DB_NAME": "tdb", "TARGET_DB_SSL_MODE": "require", "TARGET_DB_AUTH_DATABASE": "admin",
		"TARGET_DB_SSL_CA": "/ca", "TARGET_DB_SSL_CERT": "/cert", "TARGET_DB_SSL_KEY": "/key",
	})
	c, err := TargetConn(v)
	if err != nil {
		t.Fatal(err)
	}
	want := Conn{Host: "t", Port: "5433", User: "tu", Password: "tp", Database: "tdb", AuthDB: "admin", SSLMode: "require", SSLRootCert: "/ca", SSLCert: "/cert", SSLKey: "/key"}
	if c != want {
		t.Fatalf("got %+v", c)
	}
	env := c.Env()
	for _, e := range []string{"PGSSLMODE=require", "PGSSLROOTCERT=/ca", "PGSSLCERT=/cert", "PGSSLKEY=/key", "PGPASSWORD=tp", "PGHOST=t", "PGPORT=5433", "PGUSER=tu"} {
		if !slices.Contains(env, e) {
			t.Errorf("env misses %s: %v", e, env)
		}
	}
}

func TestDSNQuoting(t *testing.T) {
	c := Conn{Host: "h", Port: "5432", User: `us'er`, Password: `p'a\ss w@rd:/#?&=`, SSLMode: "require"}
	cfg, err := pgx.ParseConfig(c.DSN(`my db'x`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.User != c.User || cfg.Password != c.Password || cfg.Database != `my db'x` || cfg.Host != "h" || cfg.Port != 5432 {
		t.Fatalf("got user=%q password=%q db=%q", cfg.User, cfg.Password, cfg.Database)
	}
	if cfg.TLSConfig == nil {
		t.Fatal("sslmode require must configure TLS")
	}
	if got := ConnInfo(`a=b' -x`); got != `dbname='a=b\' -x'` {
		t.Fatal(got)
	}
}

func TestDumpArgs(t *testing.T) {
	v := values(t, map[string]string{"BACKUP_SCHEMA_ONLY": "true", "BACKUP_TABLES": "public.a, b", "BACKUP_EXCLUDE_TABLES": "logs", "PG_DUMP_FORMAT": "CUSTOM"})
	got := DumpArgs(v, "shop")
	want := []string{"-w", "-d", "dbname='shop'", "-Fc", "-Z0", "--schema-only", "-t", "public.a", "-t", "b", "--exclude-table=logs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
	v = values(t, map[string]string{"BACKUP_NO_ROLE_PASSWORDS": "1", "BACKUP_EXCLUDE_DATABASES": "_aiven,defaultdb"})
	got = DumpAllArgs(v, "postgres")
	want = []string{"-w", "-l", "postgres", "--no-role-passwords", "--exclude-database=_aiven", "--exclude-database=defaultdb"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
	// Plain dumps drop objects first so a restore can replace an existing database.
	if got := DumpArgs(values(t, nil), "shop"); !slices.Contains(got, "--clean") || !slices.Contains(got, "--if-exists") {
		t.Fatalf("plain dump without --clean: %q", got)
	}
	for _, env := range []map[string]string{{"BACKUP_DATA_ONLY": "true"}, {"PG_CLEAN": "false"}} {
		if got := DumpArgs(values(t, env), "shop"); slices.Contains(got, "--clean") {
			t.Fatalf("%v: unexpected --clean: %q", env, got)
		}
	}
	for _, a := range append(DumpArgs(v, "x"), DumpAllArgs(v, "postgres")...) {
		if strings.Contains(a, "PASSWORD") {
			t.Fatal("secret in argv")
		}
	}
}

func TestNamesAndExt(t *testing.T) {
	e := New()
	cases := []struct {
		env        map[string]string
		infix, ext string
	}{
		{map[string]string{}, "", ".sql"},
		{map[string]string{"BACKUP_SCHEMA_ONLY": "true"}, "schema", ".sql"},
		{map[string]string{"BACKUP_DATA_ONLY": "true"}, "data", ".sql"},
		{map[string]string{"BACKUP_TABLES": "a,b,c"}, "tables_3", ".sql"},
		{map[string]string{"BACKUP_TABLES": "a", "BACKUP_DATA_ONLY": "true"}, "tables_1", ".sql"},
		{map[string]string{"PG_DUMP_FORMAT": "custom"}, "", ".dump"},
		{map[string]string{"PG_DUMP_FORMAT": "custom", "BACKUP_ALL_IN_ONE": "true", "BACKUP_SCHEMA_ONLY": "true"}, "", ".sql"},
	}
	for _, c := range cases {
		tg := engine.Target{Name: "db", Values: values(t, c.env)}
		if c.env["BACKUP_ALL_IN_ONE"] != "" {
			tg.Name = AllInOneName
		}
		if got := e.NameInfix(tg); got != c.infix {
			t.Errorf("%v: infix %q, want %q", c.env, got, c.infix)
		}
		if got := e.Ext(tg); got != c.ext {
			t.Errorf("%v: ext %q, want %q", c.env, got, c.ext)
		}
	}
	if e.NativeCompression(engine.Target{}) {
		t.Fatal("pipeline compression is used for every format")
	}
}

func TestTargets(t *testing.T) {
	e := New()
	ts, err := e.Targets(t.Context(), values(t, map[string]string{"BACKUP_ALL_IN_ONE": "true"}))
	if err != nil || len(ts) != 1 || ts[0].Name != AllInOneName {
		t.Fatalf("%v %v", ts, err)
	}
	if _, err := e.Targets(t.Context(), values(t, map[string]string{})); err == nil || !strings.Contains(err.Error(), "-d flag") {
		t.Fatalf("expected the 1.x 'database name is required' error, got %v", err)
	}
	// restore of a pg_dumpall file needs no database name, and -a/-A never
	// apply to a single file.
	ts, err = e.Targets(t.Context(), values(t, map[string]string{"FILE_NAME": "x.sql", "BACKUP_ALL_DATABASES": "true"}))
	if err != nil || len(ts) != 1 || ts[0].Name != "" {
		t.Fatalf("%v %v", ts, err)
	}
	// With the command known, a stray FILE_NAME does not change a backup.
	bv := values(t, map[string]string{"FILE_NAME": "x.sql", "BACKUP_ALL_IN_ONE": "true"})
	bv.Command = "backup"
	ts, _ = e.Targets(t.Context(), bv)
	if len(ts) != 1 || ts[0].Name != AllInOneName {
		t.Fatalf("%v", ts)
	}
	rv := values(t, map[string]string{"BACKUP_ALL_IN_ONE": "true", "DB_NAME": "d"})
	rv.Command = "restore"
	ts, _ = e.Targets(t.Context(), rv)
	if len(ts) != 1 || ts[0].Name != "d" {
		t.Fatalf("%v", ts)
	}
	ts, _ = e.Targets(t.Context(), values(t, map[string]string{"DB_URL": "postgres://u:p@h/fromurl"}))
	if ts[0].Name != "fromurl" {
		t.Fatal(ts[0].Name)
	}
}

func TestConfigure(t *testing.T) {
	e := New()
	v := values(t, map[string]string{"BACKUP_SCHEMA_ONLY": "true", "BACKUP_DATA_ONLY": "true"})
	if err := e.Configure(t.Context(), v); err == nil {
		t.Fatal("schema-only + data-only must fail")
	}
	v = values(t, map[string]string{"FILE_NAME": "/backup/sub/x.sql.gz"})
	if err := e.Configure(t.Context(), v); err != nil {
		t.Fatal(err)
	}
	if v.String(config.StoragePath) != "/backup/sub" || v.String(config.FileName) != "x.sql.gz" {
		t.Fatalf("%s %s", v.String(config.StoragePath), v.String(config.FileName))
	}
	v = values(t, map[string]string{"STORAGE": "Remote"})
	_ = e.Configure(t.Context(), v)
	if v.String(config.Storage) != "sftp" {
		t.Fatal("1.x storage 'remote' must map to sftp")
	}
	v = values(t, map[string]string{"FILE_NAME": "/x.sql.gz", "STORAGE": "s3"})
	_ = e.Configure(t.Context(), v)
	if v.String(config.FileName) != "/x.sql.gz" {
		t.Fatal("remote storages keep the file as given")
	}
}

func TestSniff(t *testing.T) {
	cases := []struct {
		data, file string
		want       Kind
	}{
		{"PGDMP\x01\x0e\x00", "x.dump", KindCustom},
		{"--\n-- PostgreSQL database cluster dump\n--\n", "custom.sql", KindCluster},
		{"CREATE ROLE x;", "all_databases_20250101_000000.sql.gz", KindCluster},
		{"--\n-- PostgreSQL database dump\n--\nSET x;", "mb-upload-ab12.sql", KindPlain},
	}
	for _, c := range cases {
		k, r, err := Sniff(strings.NewReader(c.data), c.file)
		if err != nil || k != c.want {
			t.Fatalf("%q: %v %v", c.data, k, err)
		}
		b, _ := io.ReadAll(r)
		if string(b) != c.data {
			t.Fatal("Sniff must not consume data")
		}
	}
	big := strings.Repeat("INSERT INTO t VALUES (1);\n", 5000)
	_, r, err := Sniff(strings.NewReader(big), "x.sql")
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := io.ReadAll(r); string(b) != big {
		t.Fatal("large input altered")
	}
	if _, _, err := Sniff(strings.NewReader(" \n"), "x.sql"); err == nil {
		t.Fatal("empty input must fail")
	}
	boom := errors.New("boom")
	if _, _, err := Sniff(iotest.ErrReader(boom), "x.sql"); !errors.Is(err, boom) {
		t.Fatalf("read errors must surface: %v", err)
	}
}

func TestRestoreArgs(t *testing.T) {
	v := values(t, nil)
	if args := PsqlArgs(v, KindPlain, "db"); !reflect.DeepEqual(args, []string{"-X", "-q", "-w", "-d", "dbname='db'", "-v", "ON_ERROR_STOP=1", "--single-transaction"}) {
		t.Fatalf("%q", args)
	}
	if args := PsqlArgs(v, KindCustom, "db"); !reflect.DeepEqual(args, PsqlArgs(v, KindPlain, "db")) {
		t.Fatalf("custom: %q", args)
	}
	if args := PsqlArgs(v, KindCluster, "postgres"); slices.Contains(args, "--single-transaction") || slices.Contains(args, "ON_ERROR_STOP=1") {
		t.Fatalf("%q", args)
	}
	if args := ScriptArgs(v); !reflect.DeepEqual(args, []string{"-f", "-", "--no-owner", "--clean", "--if-exists"}) {
		t.Fatalf("%q", args)
	}
	v = values(t, map[string]string{"RESTORE_SINGLE_TRANSACTION": "false", "RESTORE_ON_ERROR_STOP": "true", "RESTORE_NO_OWNER": "false"})
	if args := PsqlArgs(v, KindPlain, "db"); slices.Contains(args, "--single-transaction") || !slices.Contains(args, "ON_ERROR_STOP=1") {
		t.Fatalf("%q", args)
	}
	if args := PsqlArgs(v, KindCluster, "db"); !slices.Contains(args, "ON_ERROR_STOP=1") {
		t.Fatalf("%q", args)
	}
	if args := ScriptArgs(v); slices.Contains(args, "--no-owner") {
		t.Fatalf("%q", args)
	}
}

func TestPickClients(t *testing.T) {
	avail := []int{16, 17, 18}
	cases := []struct {
		server, forced, want int
		newer                bool
	}{
		{9, 0, 16, false}, {13, 0, 16, false}, {16, 0, 16, false}, {17, 0, 17, false},
		{18, 0, 18, false}, {19, 0, 18, true}, {13, 18, 18, false}, {13, 15, 16, false},
	}
	for _, c := range cases {
		got, ok, newer := PickDump(avail, c.server, c.forced)
		if !ok || got != c.want || newer != c.newer {
			t.Errorf("server %d forced %d: got %d newer=%v", c.server, c.forced, got, newer)
		}
	}
	if _, ok, _ := PickDump(nil, 17, 0); ok {
		t.Fatal("no clients: PATH fallback")
	}
	if m, _ := PickRestore(avail, 0); m != 18 {
		t.Fatal(m)
	}
	if m, _ := PickRestore(avail, 17); m != 17 {
		t.Fatal(m)
	}
	for num, want := range map[int]int{90624: 9, 100023: 10, 130016: 13, 170002: 17, 180000: 18} {
		if ServerMajor(num) != want {
			t.Errorf("ServerMajor(%d)", num)
		}
	}
}

func TestFindClients(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"postgresql16", "postgresql18", "postgresql"} {
		_ = os.MkdirAll(filepath.Join(dir, n), 0o755)
		_ = os.WriteFile(filepath.Join(dir, n, "pg_dump"), nil, 0o755)
	}
	_ = os.MkdirAll(filepath.Join(dir, "postgresql17"), 0o755) // no binaries
	deb := filepath.Join(dir, "lib", "postgresql", "15", "bin")
	_ = os.MkdirAll(deb, 0o755)
	_ = os.WriteFile(filepath.Join(deb, "pg_dump"), nil, 0o755)
	c := FindClients(filepath.Join(dir, "postgresql*"), filepath.Join(dir, "lib", "postgresql", "*", "bin"))
	if !reflect.DeepEqual(c.Majors(), []int{15, 16, 18}) {
		t.Fatal(c.Majors())
	}
	if c.Bin(18, "psql") != filepath.Join(dir, "postgresql18", "psql") || c.Bin(14, "psql") != "psql" {
		t.Fatal(c.Bin(18, "psql"))
	}
	if len(FindClients(filepath.Join(dir, "none*")).Majors()) != 0 {
		t.Fatal("expected none")
	}
}

// TestFlags checks the 1.x flag spellings, including repeatable list flags.
func TestFlags(t *testing.T) {
	app := &cli.App{Engine: New(), Version: "test", Extend: Extend}
	root := app.Command()
	find := func(name string) *pflag.FlagSet {
		for _, c := range root.Commands() {
			if c.Name() == name {
				return c.Flags()
			}
		}
		t.Fatalf("no %s command", name)
		return nil
	}
	fs := find("backup")
	if err := fs.Parse([]string{"-t", "a", "--tables", "b,c", "--exclude-table", "x", "--exclude-db", "_aiven", "--exclude-db=defaultdb",
		"--schema-only", "--no-role-passwords", "-A", "--custom-name", "n", "--dump-format", "custom"}); err != nil {
		t.Fatal(err)
	}
	v, err := app.Registry().Resolve(config.ResolveOptions{Env: config.MapEnv{}, Flags: fs})
	if err != nil {
		t.Fatal(err)
	}
	check := map[string]string{
		Tables: "a,b,c", ExcludeTables: "x", config.ExcludeDBs: "_aiven,defaultdb", SchemaOnly: "true",
		NoRolePasswords: "true", config.AllInOne: "true", config.CustomName: "n", DumpFormat: "custom",
	}
	for k, want := range check {
		if got := v.String(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	mf := find("migrate")
	if err := mf.Parse([]string{"-I", "-a"}); err != nil {
		t.Fatal(err)
	}
	v, _ = app.Registry().Resolve(config.ResolveOptions{Env: config.MapEnv{}, Flags: mf})
	if !v.Bool(EntireInstance) || !v.Bool(config.AllDatabases) {
		t.Fatal("migrate -I -a")
	}
}

func TestEnvDocs(t *testing.T) {
	var buf bytes.Buffer
	config.WriteMarkdown(&buf, (&cli.App{Engine: New()}).Registry())
	for _, k := range []string{"DB_AUTH_DATABASE", "TARGET_DB_SSL_MODE", "TARGET_DB_AUTH_DATABASE", "PG_DUMP_FORMAT", "BACKUP_EXCLUDE_TABLES"} {
		if !strings.Contains(buf.String(), k) {
			t.Errorf("env docs miss %s", k)
		}
	}
}
