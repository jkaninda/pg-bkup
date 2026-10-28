// Package pg is the PostgreSQL engine of pg-bkup: it streams pg_dump /
// pg_dumpall output into the bkupkit pipeline and replays dumps with psql or
// pg_restore.
package pg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jkaninda/bkupkit/cli"
	"github.com/jkaninda/bkupkit/config"
	"github.com/jkaninda/bkupkit/engine"
	"github.com/jkaninda/bkupkit/execx"
)

// AllInOneName is the target (and file base) name of a pg_dumpall backup.
const AllInOneName = "all_databases"

// Engine implements engine.Engine, engine.Migrator and engine.NameSuffixer.
type Engine struct {
	// ClientGlobs are where versioned clients are looked up (ClientGlobs).
	ClientGlobs []string
}

// New returns the engine.
func New() *Engine { return &Engine{ClientGlobs: ClientGlobs} }

func (e *Engine) Tool() string { return "pg-bkup" }
func (e *Engine) Kind() string { return "postgres" }

func (e *Engine) clients() Clients { return FindClients(e.ClientGlobs...) }

func log() *slog.Logger { return slog.Default() }

func configErr(err error) error { return &cli.ExitError{Code: cli.ExitConfig, Err: err} }

// Configure validates the configuration.
func (e *Engine) Configure(_ context.Context, v *config.Values) error {
	var errs []error
	if _, err := SourceConn(v); err != nil {
		errs = append(errs, err)
	}
	if v.Has(config.TargetDBURL) || v.Has(config.TargetDBHost) {
		if _, err := TargetConn(v); err != nil {
			errs = append(errs, err)
		}
	}
	if v.Bool(SchemaOnly) && v.Bool(DataOnly) {
		errs = append(errs, errors.New("--schema-only and --data-only cannot be used together"))
	}
	normalizeStorage(v)
	localAbsoluteFile(v)
	return errors.Join(errs...)
}

// normalizeStorage maps the 1.x storage name "remote" to sftp.
func normalizeStorage(v *config.Values) {
	if v.String(config.Storage) == "remote" {
		v.Set(config.Storage, "sftp")
	}
}

// localAbsoluteFile keeps the 1.x behaviour of `restore -f /backup/x.sql.gz`
// on local storage: an absolute file path selects its directory.
func localAbsoluteFile(v *config.Values) {
	f := v.String(config.FileName)
	if !strings.HasPrefix(f, "/") || !strings.EqualFold(v.String(config.Storage), "local") {
		return
	}
	dir, base := filepath.Split(filepath.Clean(f))
	v.Set(config.StoragePath, filepath.Clean(dir))
	v.Set(config.RemotePath, "")
	v.Set(config.FileName, base)
}

// singleFile reports whether the running command works on one given file
// (restore, verify, delete), where --all-databases and --all-in-one do not
// apply. Without a known command (engine used outside the CLI), a FILE_NAME
// means the same.
func (e *Engine) singleFile(v *config.Values) bool {
	switch v.Command {
	case "":
		return v.Has(config.FileName)
	case "restore", "verify", "delete":
		return true
	}
	return false
}

// isCluster reports whether t is the pg_dumpall (--all-in-one) target.
func isCluster(t engine.Target) bool {
	return t.Name == AllInOneName && t.Values != nil && t.Values.Bool(config.AllInOne)
}

// Targets returns the databases of this run.
func (e *Engine) Targets(ctx context.Context, v *config.Values) ([]engine.Target, error) {
	normalizeStorage(v) // per-job values are resolved again from the env
	c, err := SourceConn(v)
	if err != nil {
		return nil, err
	}
	if !e.singleFile(v) {
		switch {
		case v.Bool(config.AllInOne):
			return []engine.Target{{Name: AllInOneName, Values: v}}, nil
		case v.Bool(config.AllDatabases):
			names, err := c.ListDatabases(ctx, v.List(config.ExcludeDBs))
			if err != nil {
				return nil, &cli.ExitError{Code: cli.ExitConnection, Err: fmt.Errorf("error listing databases: %w", err)}
			}
			if len(names) == 0 {
				return nil, errors.New("no databases found to back up")
			}
			log().Info("Backing up all databases", "count", len(names), "excluded", strings.Join(v.List(config.ExcludeDBs), ","))
			out := make([]engine.Target, 0, len(names))
			for _, n := range names {
				out = append(out, engine.Target{Name: n, Values: v})
			}
			return out, nil
		}
		if c.Database == "" {
			return nil, errors.New("database name is required, use DB_NAME environment variable or -d flag")
		}
	}
	return []engine.Target{{Name: c.Database, Values: v}}, nil
}

// Ext is ".dump" for the custom format, ".sql" otherwise.
func (e *Engine) Ext(t engine.Target) string {
	if !isCluster(t) && t.Values != nil && t.Values.String(DumpFormat) == FormatCustom {
		return ".dump"
	}
	return ".sql"
}

// NativeCompression is always false: the custom format is dumped with -Z0 and
// compressed by the pipeline, so every artifact is decompressed the same way
// and uploaded .dump files (already compressed internally) pass through.
func (e *Engine) NativeCompression(engine.Target) bool { return false }

// NameInfix returns "schema", "tables_N" or "data" (1.x names).
func (e *Engine) NameInfix(t engine.Target) string {
	if isCluster(t) || t.Values == nil {
		return ""
	}
	v := t.Values
	switch {
	case v.Bool(SchemaOnly):
		return "schema"
	case len(v.List(Tables)) > 0:
		return "tables_" + strconv.Itoa(len(v.List(Tables)))
	case v.Bool(DataOnly):
		return "data"
	}
	return ""
}

// dbOf is the database a target connects to.
func dbOf(t engine.Target, c Conn) string {
	if isCluster(t) || t.Name == "" {
		return c.AuthDB
	}
	return t.Name
}

func (e *Engine) Ping(ctx context.Context, t engine.Target) error {
	c, err := SourceConn(t.Values)
	if err != nil {
		return configErr(err)
	}
	if c.User == "" {
		return configErr(errors.New("missing required configuration: DB_USER (or a user in DB_URL)"))
	}
	conn, err := c.connect(ctx, dbOf(t, c))
	if err != nil {
		return err
	}
	return conn.Close(ctx)
}

func (e *Engine) ServerVersion(ctx context.Context, t engine.Target) string {
	c, err := SourceConn(t.Values)
	if err != nil {
		return ""
	}
	_, ver, err := c.ServerVersionNum(ctx, dbOf(t, c))
	if err != nil {
		return ""
	}
	return ver
}

// dumpBin picks pg_dump/pg_dumpall for the server behind c.
func (e *Engine) dumpBin(ctx context.Context, v *config.Values, c Conn, db, tool string) (string, error) {
	cl := e.clients()
	majors := cl.Majors()
	if len(majors) == 0 {
		return tool, nil
	}
	num, ver, err := c.ServerVersionNum(ctx, db)
	if err != nil {
		return "", err
	}
	major, _, newer := PickDump(majors, ServerMajor(num), v.Int(ClientVersion))
	if newer {
		log().Warn("The server is newer than every bundled PostgreSQL client; the dump may fail", "server", ver, "client", major)
	}
	bin := cl.Bin(major, tool)
	log().Info("Using PostgreSQL client", "bin", bin, "server", ver)
	return bin, nil
}

func (e *Engine) restoreBin(v *config.Values, tool string) string {
	cl := e.clients()
	if major, ok := PickRestore(cl.Majors(), v.Int(ClientVersion)); ok {
		return cl.Bin(major, tool)
	}
	return tool
}

// DumpArgs returns the pg_dump arguments for db (no connection secrets).
func DumpArgs(v *config.Values, db string) []string {
	args := []string{"-w", "-d", ConnInfo(db)}
	if v.String(DumpFormat) == FormatCustom {
		// Compression is done by the pipeline (gzip/zstd, parallel).
		args = append(args, "-Fc", "-Z0")
	}
	switch {
	case v.Bool(SchemaOnly):
		args = append(args, "--schema-only")
	case v.Bool(DataOnly):
		args = append(args, "--data-only")
	}
	// Restoring over an existing database (the common case) needs the drops;
	// custom-format dumps get them from pg_restore instead.
	if v.Bool(Clean) && !v.Bool(DataOnly) && v.String(DumpFormat) != FormatCustom {
		args = append(args, "--clean", "--if-exists")
	}
	for _, t := range v.List(Tables) {
		args = append(args, "-t", t)
	}
	for _, t := range v.List(ExcludeTables) {
		args = append(args, "--exclude-table="+t)
	}
	return args
}

// DumpAllArgs returns the pg_dumpall arguments.
func DumpAllArgs(v *config.Values, authDB string) []string {
	args := []string{"-w", "-l", authDB}
	if v.Bool(NoRolePasswords) {
		args = append(args, "--no-role-passwords")
	}
	for _, d := range v.List(config.ExcludeDBs) {
		args = append(args, "--exclude-database="+d)
	}
	return args
}

// Dump streams pg_dump (or pg_dumpall for --all-in-one) to w.
func (e *Engine) Dump(ctx context.Context, t engine.Target, w io.Writer) error {
	v := t.Values
	c, err := SourceConn(v)
	if err != nil {
		return configErr(err)
	}
	tool, args := "pg_dump", DumpArgs(v, t.Name)
	if isCluster(t) {
		tool, args = "pg_dumpall", DumpAllArgs(v, c.AuthDB)
		if v.String(DumpFormat) == FormatCustom {
			log().Warn("PG_DUMP_FORMAT=custom does not apply to --all-in-one; pg_dumpall writes plain SQL")
		}
	}
	bin, err := e.dumpBin(ctx, v, c, dbOf(t, c), tool)
	if err != nil {
		return err
	}
	return execx.Run(ctx, execx.Cmd{Name: bin, Args: args, Env: c.Env(), Stdout: w, Log: log(), StderrLevel: slog.LevelWarn})
}

// Restore detects the dump kind from its first bytes and replays it.
func (e *Engine) Restore(ctx context.Context, t engine.Target, r io.Reader, info engine.ArtifactInfo) error {
	v := t.Values
	c, err := SourceConn(v)
	if err != nil {
		return configErr(err)
	}
	kind, br, err := Sniff(r, info.File)
	if err != nil {
		return err
	}
	db := t.Name
	if db == "" {
		if kind != KindCluster {
			return configErr(errors.New("database name is required, use DB_NAME environment variable or -d flag"))
		}
		db = c.AuthDB
	}
	log().Info("Restoring", "database", db, "format", kind.String(), "client", "psql")
	psql := execx.Cmd{Name: e.restoreBin(v, "psql"), Args: PsqlArgs(v, kind, db), Env: c.Env(), Log: log()}
	switch kind {
	case KindCluster:
		// pg_dumpall creates the roles it names.
		psql.Stdin = br
		return execx.Run(ctx, psql)
	case KindCustom:
		script := execx.Cmd{Name: e.restoreBin(v, "pg_restore"), Args: ScriptArgs(v), Env: c.Env(), Stdin: br, Log: log(), StderrLevel: slog.LevelWarn}
		return pipe(ctx, script, psql, e.roleSafe(v))
	}
	in := RoleSafe(br, v.Bool(RestoreNoOwner))
	defer func() { _ = in.Close() }()
	psql.Stdin = in
	return execx.Run(ctx, psql)
}

func (e *Engine) roleSafe(v *config.Values) func(io.Reader) io.ReadCloser {
	return func(r io.Reader) io.ReadCloser { return RoleSafe(r, v.Bool(RestoreNoOwner)) }
}
