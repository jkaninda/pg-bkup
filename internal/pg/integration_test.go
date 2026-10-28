//go:build integration && unix

package pg_test

// Integration tests against real PostgreSQL servers started with the docker CLI:
//
//	go test -tags integration ./...
//
// They skip when docker is unavailable. The engine runs in process through
// bkupkit's cli.App and needs pg_dump, pg_dumpall, pg_restore and psql (18+)
// on PATH; /opt/homebrew/opt/libpq/bin and /usr/lib/postgresql/*/bin are added
// automatically. Images: PG_BKUP_MINIO_IMAGE (default pgsty/minio:latest),
// PG_BKUP_IMAGE (default jkaninda/pg-bkup:dev; the image test skips if absent).

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jkaninda/bkupkit/cli"
	"github.com/jkaninda/pg-bkup/internal/pg"
)

// password exercises quoting in env, DSN and URL handling.
const password = `p@ss:w/rd'x"y`

// Miabi's artifact regexes; the LAST match is taken as the stored artifact.
var (
	miabiDB       = regexp.MustCompile(`[\w.\-]+\.(?:sql|archive)\.gz(?:\.gpg)?`)
	miabiPlatform = regexp.MustCompile(`\.sql\.gz(\.gpg)?`)
)

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		dirs := []string{"/opt/homebrew/opt/libpq/bin", "/usr/local/opt/libpq/bin"}
		more, _ := filepath.Glob("/usr/lib/postgresql/*/bin")
		for i := len(more) - 1; i >= 0; i-- {
			dirs = append(dirs, more[i])
		}
		for _, d := range dirs {
			if _, err := os.Stat(filepath.Join(d, "pg_dump")); err == nil {
				_ = os.Setenv("PATH", d+string(os.PathListSeparator)+os.Getenv("PATH"))
				break
			}
		}
	}
	os.Exit(m.Run())
}

func requireTools(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err != nil {
		t.Skipf("docker not available: %v", err)
	}
	for _, b := range []string{"pg_dump", "pg_dumpall", "pg_restore", "psql"} {
		if _, err := exec.LookPath(b); err != nil {
			t.Skipf("%s not found on PATH", b)
		}
	}
}

func docker(t *testing.T, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	// Stdout only: "docker run" prints image pull progress on stderr, which would
	// otherwise end up in the container id on a runner without the image.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("docker %s: %v\n%s%s", strings.Join(args, " "), err, out, stderr.Bytes())
	}
	return strings.TrimSpace(string(out))
}

func image(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func publishedPort(t *testing.T, id, port string) string {
	t.Helper()
	hp := strings.Split(docker(t, "port", id, port), "\n")[0]
	_, p, err := net.SplitHostPort(hp)
	if err != nil {
		t.Fatalf("docker port %q: %v", hp, err)
	}
	return p
}

// network creates a docker network removed at the end of the test.
func network(t *testing.T) string {
	name := fmt.Sprintf("pgbkup-it-%d", time.Now().UnixNano())
	docker(t, "network", "create", name)
	t.Cleanup(func() { _ = exec.Command("docker", "network", "rm", name).Run() })
	return name
}

type server struct {
	id, port, alias string
}

var aliasSeq int

// startPG runs a PostgreSQL container with a published port, on net if given.
func startPG(t *testing.T, img, net string) *server {
	t.Helper()
	requireTools(t)
	aliasSeq++
	alias := fmt.Sprintf("pg%d", aliasSeq)
	args := []string{"run", "-d", "-p", "127.0.0.1::5432", "-e", "POSTGRES_PASSWORD=" + password}
	if net != "" {
		args = append(args, "--network", net, "--network-alias", alias)
	}
	id := docker(t, append(args, img)...)
	t.Cleanup(func() {
		if t.Failed() {
			out, _ := exec.Command("docker", "logs", "--tail", "30", id).CombinedOutput()
			t.Logf("%s logs:\n%s", img, out)
		}
		_ = exec.Command("docker", "rm", "-f", "-v", id).Run()
	})
	s := &server{id: id, port: publishedPort(t, id, "5432/tcp"), alias: alias}
	deadline := time.Now().Add(90 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		c, err := pgx.Connect(ctx, s.conn().DSN("postgres"))
		cancel()
		if err == nil {
			_ = c.Close(context.Background())
			return s
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s did not start: %v", img, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *server) conn() pg.Conn {
	return pg.Conn{Host: "127.0.0.1", Port: s.port, User: "postgres", Password: password, SSLMode: "disable"}
}

func (s *server) db(t *testing.T, db string) *pgx.Conn {
	t.Helper()
	c, err := pgx.Connect(context.Background(), s.conn().DSN(db))
	if err != nil {
		t.Fatalf("connect %s: %v", db, err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })
	return c
}

func (s *server) exec(t *testing.T, db string, sqls ...string) {
	t.Helper()
	c := s.db(t, db)
	for _, q := range sqls {
		if _, err := c.Exec(context.Background(), q); err != nil {
			t.Fatalf("%s: %s: %v", db, q, err)
		}
	}
}

func (s *server) createDB(t *testing.T, names ...string) {
	t.Helper()
	for _, n := range names {
		s.exec(t, "postgres", "CREATE DATABASE "+pgx.Identifier{n}.Sanitize())
	}
}

const seedSQL = `
CREATE TABLE users(id serial PRIMARY KEY, name text NOT NULL, email text);
INSERT INTO users(name, email) SELECT 'user' || g, 'u' || g || '@example.com' FROM generate_series(1, 3000) g;
CREATE TABLE orders(id serial PRIMARY KEY, user_id int REFERENCES users(id), total numeric(10,2), note text);
INSERT INTO orders(user_id, total, note) SELECT (g % 3000) + 1, g * 1.5, E'line\nbreak ''quoted'' \\ ü' FROM generate_series(1, 20000) g;
CREATE TABLE logs(id int, msg text);
INSERT INTO logs SELECT g, 'm' || g FROM generate_series(1, 100) g;
CREATE VIEW v_users AS SELECT name FROM users;`

func (s *server) seed(t *testing.T, db string) {
	t.Helper()
	s.createDB(t, db)
	s.exec(t, db, seedSQL)
}

// checksum describes the public tables of db: name, row count and an md5 of
// all rows in a stable order.
func (s *server) checksum(t *testing.T, db string) string {
	t.Helper()
	c := s.db(t, db)
	ctx := context.Background()
	rows, err := c.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY 1")
	if err != nil {
		t.Fatal(err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, tb := range tables {
		var n int64
		var sum string
		q := fmt.Sprintf("SELECT count(*), coalesce(md5(string_agg(t::text, '|' ORDER BY t::text)), '') FROM %s t", pgx.Identifier{tb}.Sanitize())
		if err := c.QueryRow(ctx, q).Scan(&n, &sum); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&b, "%s:%d:%s;", tb, n, sum)
	}
	return b.String()
}

func (s *server) useAsSource(t *testing.T) {
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", s.port)
	t.Setenv("DB_USERNAME", "postgres")
	t.Setenv("DB_PASSWORD", password)
}

func (s *server) useAsTarget(t *testing.T) {
	t.Setenv("TARGET_DB_HOST", "127.0.0.1")
	t.Setenv("TARGET_DB_PORT", s.port)
	t.Setenv("TARGET_DB_USERNAME", "postgres")
	t.Setenv("TARGET_DB_PASSWORD", password)
}

// isolate gives the test its own local storage and state files.
func isolate(t *testing.T) string {
	t.Helper()
	store := t.TempDir()
	for k, v := range map[string]string{
		"STORAGE_PATH": store, "HEALTH_FILE": filepath.Join(t.TempDir(), "health.json"), "TEMPLATES_DIR": "",
		"TMP_DIR": t.TempDir(), "BACKUP_CONFIG_FILE": "", "GPG_PASSPHRASE": "", "AGE_RECIPIENTS": "", "AGE_IDENTITY": "",
		"DB_NAME": "", "DB_URL": "", "STORAGE": "", "REMOTE_PATH": "", "AWS_S3_PATH": "",
	} {
		t.Setenv(k, v)
	}
	return store
}

// run executes the CLI in process and returns its combined output.
func run(t *testing.T, args ...string) (string, int) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w
	done := make(chan []byte)
	go func() {
		b, _ := io.ReadAll(r)
		done <- b
	}()
	code := (&cli.App{Engine: pg.New(), Version: "test", Extend: pg.Extend}).Run(args)
	os.Stdout, os.Stderr = oldOut, oldErr
	_ = w.Close()
	return string(<-done), code
}

func mustRun(t *testing.T, args ...string) string {
	t.Helper()
	out, code := run(t, args...)
	if code != 0 {
		t.Fatalf("%v: exit %d\n%s", args, code, out)
	}
	return out
}

var completedRe = regexp.MustCompile(`msg="Backup completed".* file=([\w.\-]+)`)

// files returns the artifact names of the "Backup completed" lines, in order.
func files(t *testing.T, out string) []string {
	t.Helper()
	var fs []string
	for _, m := range completedRe.FindAllStringSubmatch(out, -1) {
		fs = append(fs, m[1])
	}
	if len(fs) == 0 {
		t.Fatalf("no 'Backup completed' line:\n%s", out)
	}
	return fs
}

func last(fs []string) string { return fs[len(fs)-1] }

func TestIntegrationRoundTripVersions(t *testing.T) {
	for _, c := range []struct {
		img   string
		major int
	}{{"postgres:13-alpine", 13}, {"postgres:16-alpine", 16}, {"postgres:18-alpine", 18}} {
		t.Run(c.img, func(t *testing.T) {
			// With a single PATH client newer than the server, the dump
			// carries "SET transaction_timeout", which the restore guards.
			s := startPG(t, c.img, "")
			store := isolate(t)
			s.useAsSource(t)
			s.seed(t, "shop")
			want := s.checksum(t, "shop")

			// GPG passphrase (Miabi's default).
			t.Setenv("GPG_PASSPHRASE", "secret")
			f := last(files(t, mustRun(t, "backup", "-d", "shop")))
			if !regexp.MustCompile(`^shop_\d{8}_\d{6}\.sql\.gz\.gpg$`).MatchString(f) {
				t.Fatalf("file %q", f)
			}
			if _, err := os.Stat(filepath.Join(store, f)); err != nil {
				t.Fatal(err)
			}
			s.createDB(t, "shop_gpg")
			mustRun(t, "restore", "-d", "shop_gpg", "-f", f)
			if got := s.checksum(t, "shop_gpg"); got != want {
				t.Fatalf("gpg restore differs:\n got %s\nwant %s", got, want)
			}

			// age recipient, zstd.
			id, _ := age.GenerateX25519Identity()
			t.Setenv("GPG_PASSPHRASE", "")
			t.Setenv("AGE_RECIPIENTS", id.Recipient().String())
			f = last(files(t, mustRun(t, "backup", "-d", "shop", "--compression", "zstd")))
			if !strings.HasSuffix(f, ".sql.zst.age") {
				t.Fatalf("file %q", f)
			}
			t.Setenv("AGE_IDENTITY", id.String())
			s.createDB(t, "shop_age")
			mustRun(t, "restore", "-d", "shop_age", "-f", f)
			if got := s.checksum(t, "shop_age"); got != want {
				t.Fatal("age restore differs")
			}
			mustRun(t, "verify", "-d", "shop", "-f", f)
		})
	}
}

func TestIntegrationVariants(t *testing.T) {
	s := startPG(t, "postgres:17-alpine", "")
	store := isolate(t)
	s.useAsSource(t)
	s.seed(t, "shop")
	want := s.checksum(t, "shop")
	tablesOf := func(sum string) string {
		var names []string
		for _, p := range strings.Split(strings.TrimSuffix(sum, ";"), ";") {
			f := strings.Split(p, ":")
			names = append(names, f[0]+":"+f[1])
		}
		return strings.Join(names, ",")
	}

	t.Run("schema-only", func(t *testing.T) {
		f := last(files(t, mustRun(t, "backup", "-d", "shop", "--schema-only")))
		if !regexp.MustCompile(`^shop_schema_\d{8}_\d{6}\.sql\.gz$`).MatchString(f) {
			t.Fatalf("file %q", f)
		}
		s.createDB(t, "shop_schema")
		mustRun(t, "restore", "-d", "shop_schema", "-f", f)
		if got := tablesOf(s.checksum(t, "shop_schema")); got != "logs:0,orders:0,users:0" {
			t.Fatal(got)
		}
	})
	t.Run("data-only", func(t *testing.T) {
		f := last(files(t, mustRun(t, "backup", "-d", "shop", "--data-only")))
		if !strings.HasPrefix(f, "shop_data_") {
			t.Fatalf("file %q", f)
		}
		s.createDB(t, "shop_data")
		// Schema first (from a schema-only dump), then the data.
		schema := last(files(t, mustRun(t, "backup", "-d", "shop", "--schema-only")))
		mustRun(t, "restore", "-d", "shop_data", "-f", schema)
		mustRun(t, "restore", "-d", "shop_data", "-f", f)
		if got := s.checksum(t, "shop_data"); got != want {
			t.Fatal("data-only restore differs")
		}
	})
	t.Run("tables", func(t *testing.T) {
		f := last(files(t, mustRun(t, "backup", "-d", "shop", "-t", "users", "-t", "logs")))
		if !strings.HasPrefix(f, "shop_tables_2_") {
			t.Fatalf("file %q", f)
		}
		s.createDB(t, "shop_tables")
		mustRun(t, "restore", "-d", "shop_tables", "-f", f)
		if got := tablesOf(s.checksum(t, "shop_tables")); got != "logs:100,users:3000" {
			t.Fatal(got)
		}
	})
	t.Run("exclude-table", func(t *testing.T) {
		f := last(files(t, mustRun(t, "backup", "-d", "shop", "--exclude-table", "logs")))
		s.createDB(t, "shop_nologs")
		mustRun(t, "restore", "-d", "shop_nologs", "-f", f)
		if got := tablesOf(s.checksum(t, "shop_nologs")); got != "orders:20000,users:3000" {
			t.Fatal(got)
		}
	})
	t.Run("custom-format", func(t *testing.T) {
		t.Setenv("PG_DUMP_FORMAT", "custom")
		f := last(files(t, mustRun(t, "backup", "-d", "shop")))
		if !strings.HasSuffix(f, ".dump.gz") {
			t.Fatalf("file %q", f)
		}
		s.createDB(t, "shop_custom")
		mustRun(t, "restore", "-d", "shop_custom", "-f", f)
		if got := s.checksum(t, "shop_custom"); got != want {
			t.Fatal("custom restore differs")
		}
		f = last(files(t, mustRun(t, "backup", "-d", "shop", "--disable-compression")))
		if !strings.HasSuffix(f, ".dump") {
			t.Fatalf("file %q", f)
		}
		b, _ := os.ReadFile(filepath.Join(store, f))
		if !bytes.HasPrefix(b, []byte("PGDMP")) {
			t.Fatal("not a custom-format archive")
		}
	})
	t.Run("db-url-and-odd-names", func(t *testing.T) {
		odd := `we'ird db`
		s.seed(t, odd)
		u := &url.URL{Scheme: "postgres", User: url.UserPassword("postgres", password), Host: "127.0.0.1:" + s.port, Path: "/" + odd, RawQuery: "sslmode=disable"}
		t.Setenv("DB_URL", u.String())
		t.Setenv("DB_HOST", "unused.invalid")
		f := last(files(t, mustRun(t, "backup")))
		if !strings.HasPrefix(f, "we-ird-db_") {
			t.Fatalf("file %q", f)
		}
		s.createDB(t, odd+" 2")
		mustRun(t, "restore", "-d", odd+" 2", "-f", f)
		if got := s.checksum(t, odd+" 2"); got != want {
			t.Fatal("restore differs")
		}
		// JDBC form, as in 1.x.
		t.Setenv("DB_URL", fmt.Sprintf("jdbc:postgresql://127.0.0.1:%s/shop?user=postgres&password=%s", s.port, url.QueryEscape(password)))
		if fs := files(t, mustRun(t, "backup")); !strings.HasPrefix(last(fs), "shop_") {
			t.Fatal(fs)
		}
		// DB_SSL_MODE is honoured (the test server has no TLS).
		t.Setenv("DB_SSL_MODE", "require")
		if out, code := run(t, "backup"); code == 0 || !strings.Contains(strings.ToLower(out), "tls") {
			t.Fatalf("DB_SSL_MODE=require must fail against a server without TLS (exit %d)\n%s", code, out)
		}
	})
}

func TestIntegrationAllDatabases(t *testing.T) {
	s := startPG(t, "postgres:17-alpine", "")
	store := isolate(t)
	s.useAsSource(t)
	s.seed(t, "a")
	s.seed(t, "b")
	s.createDB(t, "skipme", "_aiven")
	out := mustRun(t, "backup", "--all-databases", "--exclude-db", "skipme", "--exclude-db", "_aiven")
	got := files(t, out)
	if len(got) != 2 || !strings.HasPrefix(got[0], "a_") || !strings.HasPrefix(got[1], "b_") {
		t.Fatalf("files %v\n%s", got, out)
	}
	ents, _ := os.ReadDir(store)
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), "skipme") || strings.HasPrefix(e.Name(), "postgres_") || strings.HasPrefix(e.Name(), "template") {
			t.Fatalf("unexpected %s", e.Name())
		}
	}
	// The last match of Miabi's regex is the last database's artifact.
	m := miabiDB.FindAllString(out, -1)
	if m[len(m)-1] != got[1] {
		t.Fatalf("last match %q", m[len(m)-1])
	}
}

func TestIntegrationAllInOne(t *testing.T) {
	src := startPG(t, "postgres:17-alpine", "")
	dst := startPG(t, "postgres:17-alpine", "")
	store := isolate(t)
	src.seed(t, "shop")
	src.seed(t, "crm")
	src.exec(t, "postgres", "CREATE ROLE app LOGIN PASSWORD 'x'", "GRANT CONNECT ON DATABASE shop TO app")
	src.createDB(t, "excluded")
	want := src.checksum(t, "shop")
	src.useAsSource(t)
	out := mustRun(t, "backup", "--all-in-one", "--exclude-db", "excluded", "--no-role-passwords")
	f := last(files(t, out))
	if !regexp.MustCompile(`^all_databases_\d{8}_\d{6}\.sql\.gz$`).MatchString(f) {
		t.Fatalf("file %q", f)
	}
	zr, err := gzip.NewReader(mustOpen(t, filepath.Join(store, f)))
	if err != nil {
		t.Fatal(err)
	}
	dump, _ := io.ReadAll(zr)
	if bytes.Contains(dump, []byte("SCRAM-SHA-256")) || !bytes.Contains(dump, []byte("database cluster dump")) {
		t.Fatal("pg_dumpall output with --no-role-passwords expected")
	}

	// Restore into a fresh server, without -d (the maintenance database is used).
	dst.useAsSource(t)
	mustRun(t, "restore", "-f", f)
	for _, db := range []string{"shop", "crm"} {
		if got := dst.checksum(t, db); got != want {
			t.Fatalf("%s differs after the all-in-one restore", db)
		}
	}
	var n int
	_ = dst.db(t, "postgres").QueryRow(context.Background(), "SELECT count(*) FROM pg_roles WHERE rolname = 'app'").Scan(&n)
	if n != 1 {
		t.Fatal("role app not restored")
	}
	_ = dst.db(t, "postgres").QueryRow(context.Background(), "SELECT count(*) FROM pg_database WHERE datname = 'excluded'").Scan(&n)
	if n != 0 {
		t.Fatal("excluded database restored")
	}
}

func mustOpen(t *testing.T, p string) *os.File {
	t.Helper()
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

// TestIntegrationCorruptRestore: a restore whose input fails must exit
// non-zero and leave the target database exactly as it was.
// TestIntegrationForeignRoles restores dumps naming roles the target server
// does not have (Miabi: u_<id> owners, read-only users), as a non-superuser
// owning the target database, like Miabi does.
func TestIntegrationForeignRoles(t *testing.T) {
	src := startPG(t, "postgres:17-alpine", "")
	dst := startPG(t, "postgres:16-alpine", "")
	store := isolate(t)
	src.exec(t, "postgres",
		"CREATE ROLE u_83c8ac41 LOGIN PASSWORD 'x'", "CREATE ROLE ro_83c8ac41 LOGIN PASSWORD 'x'",
		"CREATE DATABASE shop OWNER u_83c8ac41")
	src.exec(t, "shop", "SET ROLE u_83c8ac41", seedSQL,
		"CREATE FUNCTION grant_ro() RETURNS void LANGUAGE plpgsql AS $$\nBEGIN\nGRANT SELECT ON users TO ro_83c8ac41;\nEND\n$$",
		"RESET ROLE",
		"GRANT USAGE ON SCHEMA public TO ro_83c8ac41", "GRANT SELECT ON ALL TABLES IN SCHEMA public TO ro_83c8ac41",
		"ALTER DEFAULT PRIVILEGES FOR ROLE u_83c8ac41 IN SCHEMA public GRANT SELECT ON TABLES TO ro_83c8ac41")
	want := src.checksum(t, "shop")
	src.useAsSource(t)
	plain := last(files(t, mustRun(t, "backup", "-d", "shop")))
	custom := last(files(t, mustRun(t, "backup", "-d", "shop", "--dump-format", "custom")))

	dst.exec(t, "postgres", "CREATE ROLE u_other LOGIN PASSWORD '"+strings.ReplaceAll(password, "'", "''")+"'")
	dst.useAsSource(t)
	t.Setenv("DB_USERNAME", "u_other")
	for i, c := range []struct{ file, noOwner string }{{plain, "true"}, {custom, "true"}, {plain, "false"}, {custom, "false"}} {
		t.Run(fmt.Sprintf("%s no-owner=%s", filepath.Ext(strings.TrimSuffix(c.file, ".gz")), c.noOwner), func(t *testing.T) {
			db := fmt.Sprintf("copy%d", i)
			dst.exec(t, "postgres", "CREATE DATABASE "+db+" OWNER u_other")
			t.Setenv("RESTORE_NO_OWNER", c.noOwner)
			out := mustRun(t, "restore", "-d", db, "-f", c.file)
			if got := dst.checksum(t, db); got != want {
				t.Fatalf("restore differs:\n got %s\nwant %s\n%s", got, want, out)
			}
			if !strings.Contains(out, "ro_83c8ac41") {
				t.Errorf("no warning about the missing role:\n%s", out)
			}
			var owner, body string
			c := dst.db(t, db)
			if err := c.QueryRow(context.Background(), "SELECT tableowner FROM pg_tables WHERE tablename = 'users'").Scan(&owner); err != nil || owner != "u_other" {
				t.Fatalf("owner %q %v", owner, err)
			}
			if err := c.QueryRow(context.Background(), "SELECT prosrc FROM pg_proc WHERE proname = 'grant_ro'").Scan(&body); err != nil || !strings.Contains(body, "\nGRANT SELECT ON users TO ro_83c8ac41;\n") {
				t.Fatalf("function body %q %v", body, err)
			}
		})
	}

	// The same dump into a server that has the read-only role keeps its grants.
	dst.exec(t, "postgres", "CREATE ROLE ro_83c8ac41 LOGIN")
	dst.exec(t, "postgres", "CREATE DATABASE copy_ro OWNER u_other")
	mustRun(t, "restore", "-d", "copy_ro", "-f", plain)
	var ok bool
	if err := dst.db(t, "copy_ro").QueryRow(context.Background(), "SELECT has_table_privilege('ro_83c8ac41', 'users', 'SELECT')").Scan(&ok); err != nil || !ok {
		t.Fatalf("grant to an existing role lost: %v %v", ok, err)
	}

	// A genuine error still fails the restore and leaves the database as it was.
	bad := "bad_20250101_000000.sql"
	if err := os.WriteFile(filepath.Join(store, bad), []byte("CREATE TABLE t(id int);\nGRANT SELECT ON t TO nobody_here;\nSELECT * FROM no_such_table;\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dst.exec(t, "postgres", "CREATE DATABASE copy_bad OWNER u_other")
	if out, code := run(t, "restore", "-d", "copy_bad", "-f", bad); code == 0 {
		t.Fatalf("restore succeeded:\n%s", out)
	}
	if got := dst.checksum(t, "copy_bad"); got != "" {
		t.Fatalf("partial restore left %s", got)
	}
}

// foreignSource starts a server whose database posta belongs to u_src, with
// grants to a read-only role and a function whose body grants: what Miabi
// dumps look like. It returns the server and the checksum of posta.
func foreignSource(t *testing.T, img string) (*server, string) {
	t.Helper()
	src := startPG(t, img, "")
	src.exec(t, "postgres", "CREATE ROLE u_src LOGIN PASSWORD 'x'", "CREATE ROLE ro_src LOGIN PASSWORD 'x'",
		"CREATE DATABASE posta OWNER u_src")
	src.exec(t, "posta", "SET ROLE u_src", seedSQL,
		"CREATE FUNCTION grant_ro() RETURNS void LANGUAGE plpgsql AS $$\nBEGIN\nGRANT SELECT ON users TO ro_src;\nEND\n$$",
		"COMMENT ON TABLE users IS 'GRANT x TO y; ALTER TABLE t OWNER TO z;'",
		"RESET ROLE",
		"GRANT USAGE ON SCHEMA public TO ro_src", "GRANT SELECT ON ALL TABLES IN SCHEMA public TO ro_src",
		"ALTER DEFAULT PRIVILEGES FOR ROLE u_src IN SCHEMA public GRANT SELECT ON TABLES TO ro_src")
	src.exec(t, "postgres", "GRANT CONNECT ON DATABASE posta TO ro_src", "COMMENT ON DATABASE posta IS 'source'")
	return src, src.checksum(t, "posta")
}

// dumpPosta runs the server's own pg_dump (not pg-bkup) on posta.
func (s *server) dumpPosta(t *testing.T, args ...string) []byte {
	t.Helper()
	out, err := exec.Command("docker", append([]string{"exec", s.id, "pg_dump", "-U", "postgres", "-d", "posta"}, args...)...).Output()
	if err != nil {
		t.Fatalf("pg_dump %v: %v", args, err)
	}
	return out
}

// restoreUser creates a non-superuser login owning the databases it restores
// into, as Miabi's per-database users are.
func (s *server) restoreUser(t *testing.T, name string, dbs ...string) {
	t.Helper()
	s.exec(t, "postgres", "CREATE ROLE "+name+" LOGIN PASSWORD '"+strings.ReplaceAll(password, "'", "''")+"'")
	for _, db := range dbs {
		s.exec(t, "postgres", "CREATE DATABASE "+db+" OWNER "+name)
	}
	s.useAsSource(t)
	t.Setenv("DB_USERNAME", name)
}

// TestIntegrationUploads restores dumps written by other tools and servers
// (Miabi's "restore from file"): pg_dump 18 output, --create dumps, custom
// archives, INSERT dumps and hand-written SQL, into a PostgreSQL 16 database
// of another name, as a non-superuser that owns it.
func TestIntegrationUploads(t *testing.T) {
	src, want := foreignSource(t, "postgres:18-alpine")
	dst := startPG(t, "postgres:16-alpine", "")
	store := isolate(t)
	plain := src.dumpPosta(t)
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write(plain)
	_ = zw.Close()
	handWritten, err := os.ReadFile("../../migrations/init.sql")
	if err != nil {
		t.Fatal(err)
	}
	files := []struct{ name, file string }{
		{"pg_dump plain", "posta_1.sql"},
		{"pg_dump --create --clean", "posta_2.sql"},
		{"pg_dump -Fc", "posta_3.dump"},
		{"gzip", "posta_4.sql.gz"},
		{"--column-inserts", "posta_5.sql"},
	}
	for f, b := range map[string][]byte{
		"posta_1.sql": plain, "posta_2.sql": src.dumpPosta(t, "--create", "--clean", "--if-exists"),
		"posta_3.dump": src.dumpPosta(t, "-Fc"), "posta_4.sql.gz": gz.Bytes(),
		"posta_5.sql": src.dumpPosta(t, "--column-inserts"), "init.sql": handWritten,
	} {
		if err := os.WriteFile(filepath.Join(store, f), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	dbs := []string{"mail1", "mail2", "mail3", "mail4", "mail5", "handwritten"}
	dst.restoreUser(t, "mailer", dbs...)
	for i, c := range files {
		t.Run(c.name, func(t *testing.T) {
			db := dbs[i]
			out := mustRun(t, "restore", "-d", db, "-f", c.file)
			if got := dst.checksum(t, db); got != want {
				t.Fatalf("restore differs:\n got %s\nwant %s\n%s", got, want, out)
			}
			var owner string
			if err := dst.db(t, db).QueryRow(context.Background(), "SELECT tableowner FROM pg_tables WHERE tablename = 'orders'").Scan(&owner); err != nil || owner != "mailer" {
				t.Fatalf("owner %q %v", owner, err)
			}
			var comment string
			_ = dst.db(t, db).QueryRow(context.Background(), "SELECT obj_description('public.users'::regclass)").Scan(&comment)
			if comment != "GRANT x TO y; ALTER TABLE t OWNER TO z;" {
				t.Fatalf("comment %q", comment)
			}
		})
	}
	t.Run("hand-written SQL", func(t *testing.T) {
		mustRun(t, "restore", "-d", "handwritten", "-f", "init.sql")
		var n int
		if err := dst.db(t, "handwritten").QueryRow(context.Background(), "SELECT (SELECT count(*) FROM users) + (SELECT count(*) FROM orders)").Scan(&n); err != nil || n != 6 {
			t.Fatalf("rows %d %v", n, err)
		}
	})
	var n int
	if err := dst.db(t, "postgres").QueryRow(context.Background(), "SELECT count(*) FROM pg_database WHERE datname = 'posta'").Scan(&n); err != nil || n != 0 {
		t.Fatalf("the --create dump created its source database on the target (%d, %v)", n, err)
	}

	// A --create dump restored on the server it came from never drops or
	// reconnects to the source database.
	src.exec(t, "postgres", "CREATE ROLE mailer LOGIN PASSWORD '"+strings.ReplaceAll(password, "'", "''")+"'", "CREATE DATABASE posta_copy OWNER mailer")
	src.useAsSource(t)
	t.Setenv("DB_USERNAME", "mailer")
	mustRun(t, "restore", "-d", "posta_copy", "-f", "posta_2.sql")
	if got := src.checksum(t, "posta_copy"); got != want {
		t.Fatal("restore on the source server differs")
	}
	if got := src.checksum(t, "posta"); got != want {
		t.Fatal("the source database changed")
	}
}

// TestIntegrationMigrateForeignRoles migrates a database owned by roles the
// target does not have, as a non-superuser target user.
func TestIntegrationMigrateForeignRoles(t *testing.T) {
	src, want := foreignSource(t, "postgres:17-alpine")
	dst := startPG(t, "postgres:16-alpine", "")
	isolate(t)
	dst.restoreUser(t, "mailer", "mail")
	src.useAsSource(t)
	dst.useAsTarget(t)
	t.Setenv("TARGET_DB_USERNAME", "mailer")
	t.Setenv("TARGET_DB_NAME", "mail")
	mustRun(t, "migrate", "-d", "posta")
	if got := dst.checksum(t, "mail"); got != want {
		t.Fatalf("migrated database differs:\n got %s\nwant %s", got, want)
	}
}

func TestIntegrationCorruptRestore(t *testing.T) {
	s := startPG(t, "postgres:17-alpine", "")
	store := isolate(t)
	s.useAsSource(t)
	s.seed(t, "shop")
	t.Setenv("BACKUP_INDEX", "false") // corrupted copies have no index entry anyway
	good := last(files(t, mustRun(t, "backup", "-d", "shop")))
	raw, _ := os.ReadFile(filepath.Join(store, good))
	zr, _ := gzip.NewReader(bytes.NewReader(raw))
	plain, _ := io.ReadAll(zr)

	s.createDB(t, "victim")
	s.exec(t, "victim", "CREATE TABLE marker(v text)", "INSERT INTO marker VALUES ('before')")
	before := s.checksum(t, "victim")

	write := func(name string, b []byte) string {
		if err := os.WriteFile(filepath.Join(store, name), b, 0o600); err != nil {
			t.Fatal(err)
		}
		return name
	}
	flipped := append([]byte(nil), raw...)
	for i := len(flipped) / 2; i < len(flipped)/2+64; i++ {
		flipped[i] ^= 0xff
	}
	t.Setenv("GPG_PASSPHRASE", "secret")
	enc := last(files(t, mustRun(t, "backup", "-d", "shop")))
	encRaw, _ := os.ReadFile(filepath.Join(store, enc))
	encTrunc := encRaw[:len(encRaw)*3/4]

	cases := map[string]string{
		"truncated gzip":     write("trunc_20250101_000000.sql.gz", raw[:len(raw)*2/3]),
		"corrupted gzip":     write("flip_20250101_000000.sql.gz", flipped),
		"bad SQL at the end": write("bad_20250101_000000.sql", append(append([]byte(nil), plain...), []byte("\nSELECT * FROM no_such_table;\n")...)),
		"truncated gpg":      write("enc_20250101_000000.sql.gz.gpg", encTrunc),
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			out, code := run(t, "restore", "-d", "victim", "-f", f)
			if code == 0 {
				t.Fatalf("restore succeeded\n%s", out)
			}
			if got := s.checksum(t, "victim"); got != before {
				t.Fatalf("victim changed:\n got %s\nwant %s\n%s", got, before, out)
			}
		})
	}
}

func TestIntegrationMigrate(t *testing.T) {
	net := ""
	src := startPG(t, "postgres:13-alpine", net)
	dst := startPG(t, "postgres:17-alpine", net)
	isolate(t)
	src.seed(t, "shop")
	src.seed(t, "crm")
	want := src.checksum(t, "shop")
	src.useAsSource(t)
	dst.useAsTarget(t)

	t.Run("one database", func(t *testing.T) {
		t.Setenv("TARGET_DB_NAME", "shop_copy")
		t.Setenv("TARGET_DB_SSL_MODE", "disable")
		t.Setenv("TARGET_DB_AUTH_DATABASE", "postgres")
		mustRun(t, "migrate", "-d", "shop")
		if got := dst.checksum(t, "shop_copy"); got != want {
			t.Fatal("migrated database differs")
		}
		// TARGET_DB_SSL_MODE reaches the target client (no TLS on the server).
		t.Setenv("TARGET_DB_SSL_MODE", "require")
		t.Setenv("TARGET_DB_NAME", "shop_tls")
		if out, code := run(t, "migrate", "-d", "shop"); code == 0 {
			t.Fatalf("TARGET_DB_SSL_MODE=require must fail\n%s", out)
		}
		// A target auth database that does not exist is used (and fails).
		t.Setenv("TARGET_DB_SSL_MODE", "")
		t.Setenv("TARGET_DB_AUTH_DATABASE", "nope")
		if out, code := run(t, "migrate", "-d", "shop"); code == 0 || !strings.Contains(out, "nope") {
			t.Fatalf("TARGET_DB_AUTH_DATABASE ignored (exit %d)\n%s", code, out)
		}
	})
	t.Run("all databases", func(t *testing.T) {
		mustRun(t, "migrate", "--all-databases")
		for _, db := range []string{"shop", "crm"} {
			if got := dst.checksum(t, db); got != want {
				t.Fatalf("%s differs", db)
			}
		}
	})
	t.Run("entire instance", func(t *testing.T) {
		fresh := startPG(t, "postgres:17-alpine", "")
		fresh.useAsTarget(t)
		src.exec(t, "postgres", "CREATE ROLE reporting LOGIN")
		// pg13 stores md5 password hashes that a pg17 scram-sha-256 hba
		// rejects: keep the target's passwords.
		mustRun(t, "migrate", "-I", "--no-role-passwords")
		if got := fresh.checksum(t, "crm"); got != want {
			t.Fatal("instance migration differs")
		}
		var n int
		_ = fresh.db(t, "postgres").QueryRow(context.Background(), "SELECT count(*) FROM pg_roles WHERE rolname = 'reporting'").Scan(&n)
		if n != 1 {
			t.Fatal("role not migrated")
		}
	})
}

type minio struct {
	endpoint, alias, bucket string
	client                  *s3.Client
}

func startMinIO(t *testing.T, net string) *minio {
	t.Helper()
	args := []string{"run", "-d", "-p", "127.0.0.1::9000", "-e", "MINIO_ROOT_USER=minioadmin", "-e", "MINIO_ROOT_PASSWORD=minioadmin"}
	if net != "" {
		args = append(args, "--network", net, "--network-alias", "minio")
	}
	id := docker(t, append(args, image("PG_BKUP_MINIO_IMAGE", "pgsty/minio:latest"), "server", "/data")...)
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", "-v", id).Run() })
	m := &minio{endpoint: "127.0.0.1:" + publishedPort(t, id, "9000/tcp"), alias: "minio:9000", bucket: "backups"}
	m.client = s3.New(s3.Options{
		Region: "us-east-1", BaseEndpoint: aws.String("http://" + m.endpoint), UsePathStyle: true,
		Credentials: credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", ""),
	})
	deadline := time.Now().Add(60 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := m.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(m.bucket)})
		cancel()
		if err == nil {
			return m
		}
		if time.Now().After(deadline) {
			t.Fatalf("create bucket: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// miabiEnv is the environment Miabi sets for pg-bkup (PLAN §2).
func (m *minio) miabiEnv(endpoint string) map[string]string {
	return map[string]string{
		"AWS_S3_ENDPOINT": endpoint, "AWS_S3_BUCKET_NAME": m.bucket, "AWS_REGION": "us-east-1",
		"AWS_ACCESS_KEY": "minioadmin", "AWS_SECRET_KEY": "minioadmin",
		"AWS_DISABLE_SSL": "true", "AWS_FORCE_PATH_STYLE": "true", "GPG_PASSPHRASE": "miabi-passphrase",
	}
}

func (m *minio) exists(key string) bool {
	_, err := m.client.HeadObject(context.Background(), &s3.HeadObjectInput{Bucket: aws.String(m.bucket), Key: aws.String(key)})
	return err == nil
}

// lastArtifact applies Miabi's rules: the last match of the DB regex is the
// artifact, and the last platform-regex match is that same artifact's suffix.
func lastArtifact(t *testing.T, out string) string {
	t.Helper()
	all := miabiDB.FindAllStringIndex(out, -1)
	if len(all) == 0 {
		t.Fatalf("no artifact in output\n%s", out)
	}
	lastDB := all[len(all)-1]
	plat := miabiPlatform.FindAllStringIndex(out, -1)
	if len(plat) == 0 || plat[len(plat)-1][1] != lastDB[1] {
		t.Fatalf("platform regex last match does not end the artifact name\n%s", out)
	}
	return out[lastDB[0]:lastDB[1]]
}

func TestIntegrationMiabiContract(t *testing.T) {
	s := startPG(t, "postgres:17-alpine", "")
	m := startMinIO(t, "")
	store := isolate(t)
	s.useAsSource(t)
	s.seed(t, "appdb")
	want := s.checksum(t, "appdb")
	for k, v := range m.miabiEnv(m.endpoint) {
		t.Setenv(k, v)
	}

	// Local: backup -d db.
	t.Setenv("BACKUP_RETENTION_COUNT", "1")
	mustRun(t, "backup", "-d", "appdb")
	time.Sleep(1100 * time.Millisecond)
	out := mustRun(t, "backup", "-d", "appdb")
	local := lastArtifact(t, out)
	if !strings.HasSuffix(local, ".sql.gz.gpg") || local != last(files(t, out)) {
		t.Fatalf("artifact %q\n%s", local, out)
	}
	if _, err := os.Stat(filepath.Join(store, local)); err != nil {
		t.Fatalf("last match is not the stored file: %v", err)
	}
	s.createDB(t, "appdb_local")
	mustRun(t, "restore", "-d", "appdb_local", "-f", local)
	if s.checksum(t, "appdb_local") != want {
		t.Fatal("local restore differs")
	}

	// S3: backup --storage s3 -d db --path p; restore -d db -f file --storage s3 --path p.
	out = mustRun(t, "backup", "--storage", "s3", "-d", "appdb", "--path", "miabi/pg")
	remote := lastArtifact(t, out)
	if !m.exists("miabi/pg/" + remote) {
		t.Fatalf("last match %q is not the stored object\n%s", remote, out)
	}
	s.createDB(t, "appdb_s3")
	mustRun(t, "restore", "-d", "appdb_s3", "-f", remote, "--storage", "s3", "--path", "miabi/pg")
	if s.checksum(t, "appdb_s3") != want {
		t.Fatal("s3 restore differs")
	}

	// Uploaded files: mb-upload-<hex>.sql, .sql.gz and .dump (made by pg_dump
	// itself, not by pg-bkup), restored with restore -d db -f file.
	t.Setenv("GPG_PASSPHRASE", "")
	t.Setenv("BACKUP_RETENTION_COUNT", "")
	pgdump := func(args ...string) []byte {
		cmd := exec.Command("pg_dump", append([]string{"-h", "127.0.0.1", "-p", s.port, "-U", "postgres", "-d", "appdb"}, args...)...)
		cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
		b, err := cmd.Output()
		if err != nil {
			t.Fatalf("pg_dump: %v", err)
		}
		return b
	}
	plain := pgdump()
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write(plain)
	_ = zw.Close()
	dumpFile := filepath.Join(t.TempDir(), "x.dump")
	pgdump("-Fc", "-f", dumpFile) // seekable output, as a user upload would be
	custom, _ := os.ReadFile(dumpFile)
	for i, up := range []struct {
		ext  string
		data []byte
	}{{".sql", plain}, {".sql.gz", gz.Bytes()}, {".dump", custom}} {
		name := fmt.Sprintf("mb-upload-%08x%s", 0xa1b2c3d0+i, up.ext)
		if err := os.WriteFile(filepath.Join(store, name), up.data, 0o600); err != nil {
			t.Fatal(err)
		}
		db := fmt.Sprintf("upload_%d", i)
		s.createDB(t, db)
		mustRun(t, "restore", "-d", db, "-f", name)
		if s.checksum(t, db) != want {
			t.Fatalf("%s restore differs", name)
		}
	}
}

// TestIntegrationImage runs the built image like Miabi does and checks the
// client major chosen per server version.
func TestIntegrationImage(t *testing.T) {
	requireTools(t)
	img := image("PG_BKUP_IMAGE", "jkaninda/pg-bkup:dev")
	if err := exec.Command("docker", "image", "inspect", img).Run(); err != nil {
		t.Skipf("image %s not built", img)
	}
	net := network(t)
	m := startMinIO(t, net)
	for _, c := range []struct {
		img    string
		client string
	}{{"postgres:13-alpine", "16"}, {"postgres:17-alpine", "17"}, {"postgres:18-alpine", "18"}} {
		t.Run(c.img, func(t *testing.T) {
			s := startPG(t, c.img, net)
			s.seed(t, "appdb")
			want := s.checksum(t, "appdb")
			envArgs := []string{"-e", "DB_HOST=" + s.alias, "-e", "DB_PORT=5432", "-e", "DB_USERNAME=postgres", "-e", "DB_PASSWORD=" + password, "-e", "LOG_LEVEL=debug"}
			for k, v := range m.miabiEnv(m.alias) {
				envArgs = append(envArgs, "-e", k+"="+v)
			}
			base := append([]string{"run", "--rm", "--network", net}, envArgs...)
			out := docker(t, append(base, img, "backup", "--storage", "s3", "-d", "appdb", "--path", "img")...)
			if !regexp.MustCompile(`postgresql` + c.client + `/pg_dump`).MatchString(out) {
				t.Fatalf("expected client %s\n%s", c.client, out)
			}
			f := lastArtifact(t, out)
			if !m.exists("img/" + f) {
				t.Fatalf("%q not stored", f)
			}
			s.createDB(t, "appdb_r")
			docker(t, append(base, img, "restore", "-d", "appdb_r", "-f", f, "--storage", "s3", "--path", "img")...)
			if s.checksum(t, "appdb_r") != want {
				t.Fatal("restore differs")
			}
		})
	}
	docker(t, "run", "--rm", "--entrypoint", "/bin/sh", img, "-c", "rm -f /backup/x && pg-bkup version && tini --version")
}
