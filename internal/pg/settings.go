package pg

import (
	"slices"
	"strings"

	"github.com/jkaninda/bkupkit/cli"
	"github.com/jkaninda/bkupkit/config"
	"github.com/jkaninda/bkupkit/engine"
	"github.com/spf13/cobra"
)

// pg-bkup specific settings.
const (
	AuthDatabase       = "DB_AUTH_DATABASE"
	TargetAuthDatabase = "TARGET_DB_AUTH_DATABASE"
	TargetSSLCert      = "TARGET_DB_SSL_CERT"
	TargetSSLKey       = "TARGET_DB_SSL_KEY"
	DumpFormat         = "PG_DUMP_FORMAT"
	ClientVersion      = "PG_CLIENT_VERSION"
	SchemaOnly         = "BACKUP_SCHEMA_ONLY"
	DataOnly           = "BACKUP_DATA_ONLY"
	Tables             = "BACKUP_TABLES"
	ExcludeTables      = "BACKUP_EXCLUDE_TABLES"
	NoRolePasswords    = "BACKUP_NO_ROLE_PASSWORDS"
	EntireInstance     = "MIGRATE_ENTIRE_INSTANCE"
	SingleTransaction  = "RESTORE_SINGLE_TRANSACTION"
	OnErrorStop        = "RESTORE_ON_ERROR_STOP"
	Clean              = "PG_CLEAN"
	RestoreNoOwner     = "RESTORE_NO_OWNER"
)

// PG_DUMP_FORMAT values.
const (
	FormatPlain  = "plain"
	FormatCustom = "custom"
)

// sslModes are the libpq sslmode values.
var sslModes = []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}

// Settings registers the database settings and the pg-bkup options.
func (e *Engine) Settings(r *config.Registry) {
	config.RegisterDatabase(r, "5432")
	backup := []string{"backup"}
	// 1.x accepted "remote" for SFTP; it is mapped to sftp in Configure/Targets.
	r.Override(config.Storage, func(s *config.Setting) {
		if !slices.Contains(s.Enum, "remote") {
			s.Enum = append(s.Enum, "remote")
		}
	})
	r.Override(config.DBURL, func(s *config.Setting) {
		s.Usage = "Connection URL: postgres://user:pass@host:5432/db?sslmode=require or jdbc:postgresql://host:5432/db?user=u&password=p (overrides host, port, user, password; its database is used when DB_NAME is not set)"
	})
	r.Override(config.DBSSLMode, func(s *config.Setting) {
		s.Enum = sslModes
		s.Usage = "libpq sslmode, passed to pg_dump, psql and pg_restore as PGSSLMODE (empty = libpq default, prefer)"
	})
	r.Override(config.DBSSLCA, func(s *config.Setting) { s.Usage = "CA certificate file (PGSSLROOTCERT)" })
	r.Override(config.DBSSLCert, func(s *config.Setting) { s.Usage = "Client certificate file (PGSSLCERT)" })
	r.Override(config.DBSSLKey, func(s *config.Setting) { s.Usage = "Client key file (PGSSLKEY)" })
	r.Override(config.AllDatabases, func(s *config.Setting) {
		s.Usage = "backup: every database (templates and 'postgres' excluded), one file each; migrate: every database into the same name on the target"
	})
	r.Override(config.AllInOne, func(s *config.Setting) {
		s.Usage = "Back up the whole instance (roles, tablespaces, all databases) into one file with pg_dumpall"
	})
	r.Override(config.ExcludeDBs, func(s *config.Setting) {
		s.Usage = "Databases to skip with --all-databases / --all-in-one, comma separated (repeatable flag)"
	})
	r.Override(config.TargetDBSSL, func(s *config.Setting) {
		s.Enum = sslModes
		s.Usage = "migrate: target libpq sslmode"
	})
	r.Override(config.TargetDBURL, func(s *config.Setting) {
		s.Usage = "migrate: target connection URL (postgres:// or jdbc:postgresql://)"
	})
	r.Add(
		config.Setting{Key: AuthDatabase, Default: "postgres", Group: config.GroupDatabase, Usage: "Maintenance database used to list databases, create databases and run pg_dumpall"},
		config.Setting{Key: TargetAuthDatabase, Default: "postgres", Group: config.GroupDatabase, Usage: "migrate: target maintenance database"},
		config.Setting{Key: TargetSSLCert, Group: config.GroupDatabase, Usage: "migrate: target client certificate file"},
		config.Setting{Key: TargetSSLKey, Group: config.GroupDatabase, Usage: "migrate: target client key file"},
		config.Setting{Key: DumpFormat, Default: FormatPlain, Enum: []string{FormatPlain, FormatCustom}, Flag: "dump-format", Commands: backup, Group: config.GroupDatabase,
			Usage: "Dump format: plain (.sql, psql) or custom (.dump, pg_dump -Fc -Z0, pg_restore); COMPRESSION applies to both"},
		config.Setting{Key: ClientVersion, Kind: config.Int, Group: config.GroupDatabase, Usage: "Force a PostgreSQL client major (16, 17, 18); default: the lowest bundled client >= the server version"},
		config.Setting{Key: SchemaOnly, Kind: config.Bool, Flag: "schema-only", Commands: backup, Group: config.GroupDatabase, Usage: "Dump the schema only (file name infix _schema_)"},
		config.Setting{Key: DataOnly, Kind: config.Bool, Flag: "data-only", Commands: backup, Group: config.GroupDatabase, Usage: "Dump the data only (file name infix _data_)"},
		config.Setting{Key: Tables, Kind: config.List, Flag: "tables", Short: "t", Commands: backup, Group: config.GroupDatabase, Usage: "Only these tables (pg_dump -t patterns), comma separated or repeated (file name infix _tables_N_)"},
		config.Setting{Key: ExcludeTables, Kind: config.List, Flag: "exclude-table", Commands: backup, Group: config.GroupDatabase, Usage: "Skip these tables (pg_dump --exclude-table patterns), comma separated or repeated"},
		config.Setting{Key: NoRolePasswords, Kind: config.Bool, Flag: "no-role-passwords", Commands: []string{"backup", "migrate"}, Group: config.GroupDatabase, Usage: "--all-in-one / migrate --entire-instance: do not dump role passwords (managed services such as Aiven, OVH)"},
		config.Setting{Key: EntireInstance, Kind: config.Bool, Flag: "entire-instance", Short: "I", Commands: []string{"migrate"}, Group: config.GroupDatabase, Usage: "migrate: copy the whole instance (roles, tablespaces, all databases) with pg_dumpall"},
		config.Setting{Key: SingleTransaction, Kind: config.Bool, Default: "true", Group: config.GroupDatabase, Usage: "Restore a database dump in one transaction (all or nothing); not used for --all-in-one dumps"},
		config.Setting{Key: OnErrorStop, Default: "auto", Enum: []string{"auto", "true", "false"}, Group: config.GroupDatabase,
			Usage: "Stop a psql restore at the first error; auto = on for database dumps, off for --all-in-one (pg_dumpall) dumps, which always contain 'role already exists' errors"},
		config.Setting{Key: Clean, Kind: config.Bool, Default: "true", Group: config.GroupDatabase,
			Usage: "Drop objects before recreating them (pg_dump/pg_restore --clean --if-exists), so a restore replaces an existing database inside its transaction"},
		config.Setting{Key: RestoreNoOwner, Kind: config.Bool, Default: "true", Group: config.GroupDatabase, Usage: "pg_restore (custom format): do not restore object ownership (--no-owner)"},
	)
}

// listValue is a repeatable, comma-separated string flag ("-t a -t b,c")
// whose String() is a plain comma list, as the config registry expects.
type listValue struct{ items []string }

func (l *listValue) String() string { return strings.Join(l.items, ",") }
func (l *listValue) Type() string   { return "strings" }
func (l *listValue) Set(s string) error {
	l.items = append(l.items, config.SplitList(s)...)
	return nil
}

// Extend registers the 1.x repeatable list flags before bkupkit binds the
// registry flags (BindFlags skips flags that already exist).
func Extend(root *cobra.Command, _ *cli.App) {
	for _, c := range root.Commands() {
		if c.Name() != "backup" {
			continue
		}
		fs := c.Flags()
		fs.VarP(&listValue{}, "tables", "t", "Only these tables (pg_dump -t patterns), comma separated or repeated (env "+Tables+")")
		fs.Var(&listValue{}, "exclude-table", "Skip these tables, comma separated or repeated (env "+ExcludeTables+")")
		fs.Var(&listValue{}, "exclude-db", "Databases to skip with --all-databases / --all-in-one, comma separated or repeated (env "+config.ExcludeDBs+")")
	}
}

var _ engine.Engine = (*Engine)(nil)
