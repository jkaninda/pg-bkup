package pg

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jkaninda/bkupkit/config"
)

// Conn is a resolved PostgreSQL connection.
type Conn struct {
	Host     string
	Port     string
	User     string
	Password string
	// Database is the database to back up / restore into ("" = none given).
	Database string
	// AuthDB is the maintenance database (listing, CREATE DATABASE, pg_dumpall).
	AuthDB      string
	SSLMode     string
	SSLRootCert string
	SSLCert     string
	SSLKey      string
}

// keys selects the source (DB_*) or target (TARGET_DB_*) settings.
type keys struct {
	url, host, port, name, user, pass, ssl, ca, cert, key, auth string
}

var (
	sourceKeys = keys{config.DBURL, config.DBHost, config.DBPort, config.DBName, config.DBUser, config.DBPassword,
		config.DBSSLMode, config.DBSSLCA, config.DBSSLCert, config.DBSSLKey, AuthDatabase}
	targetKeys = keys{config.TargetDBURL, config.TargetDBHost, config.TargetDBPort, config.TargetDBName, config.TargetDBUser, config.TargetDBPass,
		config.TargetDBSSL, config.TargetDBSSLCA, TargetSSLCert, TargetSSLKey, TargetAuthDatabase}
)

// SourceConn returns the DB_* connection.
func SourceConn(v *config.Values) (Conn, error) { return connFrom(v, sourceKeys) }

// TargetConn returns the TARGET_DB_* connection (migrate).
func TargetConn(v *config.Values) (Conn, error) { return connFrom(v, targetKeys) }

func connFrom(v *config.Values, k keys) (Conn, error) {
	c := Conn{
		Host: v.String(k.host), Port: v.String(k.port), User: v.String(k.user), Password: v.String(k.pass),
		Database: v.String(k.name), AuthDB: v.String(k.auth), SSLMode: v.String(k.ssl),
		SSLRootCert: v.String(k.ca), SSLCert: v.String(k.cert), SSLKey: v.String(k.key),
	}
	if raw := v.String(k.url); raw != "" {
		u, err := ParseURL(raw)
		if err != nil {
			return c, fmt.Errorf("%s: %w", k.url, err)
		}
		// The URL overrides host, port, user and password. Its database is used
		// unless a database was given explicitly (-d / DB_NAME); TLS settings
		// from the URL win over the defaults but not over explicit env.
		c.Host, c.User, c.Password = u.Host, u.User, u.Password
		if u.Port != "" {
			c.Port = u.Port
		}
		if !v.IsSet(k.name) {
			c.Database = u.Database
		}
		pick := func(dst *string, key, val string) {
			if val != "" && !v.IsSet(key) {
				*dst = val
			}
		}
		pick(&c.SSLMode, k.ssl, u.SSLMode)
		pick(&c.SSLRootCert, k.ca, u.SSLRootCert)
		pick(&c.SSLCert, k.cert, u.SSLCert)
		pick(&c.SSLKey, k.key, u.SSLKey)
	}
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == "" {
		c.Port = "5432"
	}
	if c.AuthDB == "" {
		c.AuthDB = "postgres"
	}
	if c.SSLMode != "" && !validSSLMode(c.SSLMode) {
		return c, fmt.Errorf("invalid sslmode %q (use %s)", c.SSLMode, strings.Join(sslModes, ", "))
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return c, fmt.Errorf("invalid port %q", c.Port)
	}
	return c, nil
}

func validSSLMode(m string) bool {
	for _, s := range sslModes {
		if s == m {
			return true
		}
	}
	return false
}

// URL is a parsed DB_URL.
type URL struct {
	Host, Port, User, Password, Database  string
	SSLMode, SSLRootCert, SSLCert, SSLKey string
}

// ParseURL accepts postgres://, postgresql:// and jdbc:postgresql:// URLs.
// Credentials may be in the userinfo (percent-encoded) or, as in JDBC, in the
// user and password query parameters.
func ParseURL(raw string) (URL, error) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "jdbc:")
	u, err := url.Parse(s)
	if err != nil {
		return URL{}, errors.New("cannot parse the URL (percent-encode special characters in the user and password)")
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return URL{}, fmt.Errorf("unsupported scheme %q (use postgres://, postgresql:// or jdbc:postgresql://)", u.Scheme)
	}
	q := u.Query()
	out := URL{
		Host: u.Hostname(), Port: u.Port(),
		Database:    strings.TrimPrefix(u.Path, "/"),
		SSLMode:     strings.ToLower(q.Get("sslmode")),
		SSLRootCert: q.Get("sslrootcert"), SSLCert: q.Get("sslcert"), SSLKey: q.Get("sslkey"),
	}
	if u.User != nil {
		out.User = u.User.Username()
		out.Password, _ = u.User.Password()
	}
	if v := q.Get("user"); v != "" {
		out.User = v
	}
	if v := q.Get("password"); v != "" {
		out.Password = v
	}
	// JDBC spells it ssl=true / sslmode; map ssl=true to require.
	if out.SSLMode == "" && strings.EqualFold(q.Get("ssl"), "true") {
		out.SSLMode = "require"
	}
	if out.Host == "" {
		return out, errors.New("the URL has no host")
	}
	return out, nil
}

// Env returns the libpq environment for a client process. Secrets go here,
// never into argv. The database is passed separately with ConnInfo.
func (c Conn) Env() []string {
	env := []string{
		"PGHOST=" + c.Host,
		"PGPORT=" + c.Port,
		"PGUSER=" + c.User,
		"PGAPPNAME=pg-bkup",
		"PGCONNECT_TIMEOUT=30",
	}
	if c.Password != "" {
		env = append(env, "PGPASSWORD="+c.Password)
	}
	if c.SSLMode != "" {
		env = append(env, "PGSSLMODE="+c.SSLMode)
	}
	if c.SSLRootCert != "" {
		env = append(env, "PGSSLROOTCERT="+c.SSLRootCert)
	}
	if c.SSLCert != "" {
		env = append(env, "PGSSLCERT="+c.SSLCert)
	}
	if c.SSLKey != "" {
		env = append(env, "PGSSLKEY="+c.SSLKey)
	}
	return env
}

// quote quotes a libpq keyword/value connection string value.
func quote(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
}

// ConnInfo returns "dbname='db'" for the -d argument of pg_dump, psql and
// pg_restore. Passing a conninfo (not a bare name) means a database name
// containing '=' or starting with '-' is never misinterpreted.
func ConnInfo(db string) string { return "dbname=" + quote(db) }

// DSN returns a keyword/value connection string for pgx, with every value
// quoted (no URL escaping problems with special characters).
func (c Conn) DSN(db string) string {
	parts := []string{
		"host=" + quote(c.Host), "port=" + quote(c.Port), "user=" + quote(c.User),
		"dbname=" + quote(db), "application_name='pg-bkup'", "connect_timeout='30'",
	}
	if c.Password != "" {
		parts = append(parts, "password="+quote(c.Password))
	}
	if c.SSLMode != "" {
		parts = append(parts, "sslmode="+quote(c.SSLMode))
	}
	if c.SSLRootCert != "" {
		parts = append(parts, "sslrootcert="+quote(c.SSLRootCert))
	}
	if c.SSLCert != "" {
		parts = append(parts, "sslcert="+quote(c.SSLCert))
	}
	if c.SSLKey != "" {
		parts = append(parts, "sslkey="+quote(c.SSLKey))
	}
	return strings.Join(parts, " ")
}

// connect opens a pgx connection to db.
func (c Conn) connect(ctx context.Context, db string) (*pgx.Conn, error) {
	cfg, err := pgx.ParseConfig(c.DSN(db))
	if err != nil {
		return nil, fmt.Errorf("connection settings: %w", err)
	}
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL %s:%s/%s: %w", c.Host, c.Port, db, err)
	}
	return conn, nil
}

// ServerVersionNum returns server_version_num (e.g. 170002) and server_version.
func (c Conn) ServerVersionNum(ctx context.Context, db string) (int, string, error) {
	conn, err := c.connect(ctx, db)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()
	var num, ver string
	if err := conn.QueryRow(ctx, "SELECT current_setting('server_version_num'), current_setting('server_version')").Scan(&num, &ver); err != nil {
		return 0, "", fmt.Errorf("query server version: %w", err)
	}
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0, ver, fmt.Errorf("unexpected server_version_num %q", num)
	}
	return n, ver, nil
}

// ListDatabases returns the non-template databases except "postgres" and
// the excluded names, from the maintenance database.
func (c Conn) ListDatabases(ctx context.Context, exclude []string) ([]string, error) {
	conn, err := c.connect(ctx, c.AuthDB)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()
	rows, err := conn.Query(ctx, `SELECT datname FROM pg_database
		WHERE datistemplate = false AND datallowconn AND datname NOT IN ('postgres', 'template0', 'template1')
		ORDER BY datname`)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	return filterExcluded(names, exclude), nil
}

func filterExcluded(names, exclude []string) []string {
	skip := map[string]bool{}
	for _, e := range exclude {
		skip[e] = true
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		if !skip[n] {
			out = append(out, n)
		}
	}
	return out
}

// EnsureDatabase creates db (from the maintenance database) if it is missing.
func (c Conn) EnsureDatabase(ctx context.Context, db string) (bool, error) {
	conn, err := c.connect(ctx, c.AuthDB)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()
	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", db).Scan(&exists); err != nil {
		return false, fmt.Errorf("check database %q: %w", db, err)
	}
	if exists {
		return false, nil
	}
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{db}.Sanitize()); err != nil {
		return false, fmt.Errorf("create database %q: %w", db, err)
	}
	return true, nil
}
