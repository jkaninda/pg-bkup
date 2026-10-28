package pg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/jkaninda/bkupkit/config"
	"github.com/jkaninda/bkupkit/engine"
	"github.com/jkaninda/bkupkit/execx"
)

// InstanceName is the source/target name of an --entire-instance migration.
const InstanceName = "(instance)"

var _ engine.Migrator = (*Engine)(nil)

// MigrateTargets returns the source/target pairs: one database
// (DB_NAME -> TARGET_DB_NAME), every database (-a, same names), or the
// whole instance (-I).
func (e *Engine) MigrateTargets(ctx context.Context, v *config.Values) ([][2]engine.Target, error) {
	src, err := SourceConn(v)
	if err != nil {
		return nil, err
	}
	dst, err := TargetConn(v)
	if err != nil {
		return nil, err
	}
	var missing []string
	if !v.Has(config.TargetDBURL) {
		for _, k := range []string{config.TargetDBHost, config.TargetDBUser} {
			if !v.Has(k) {
				missing = append(missing, k)
			}
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required configuration for the target database: %v", missing)
	}
	pair := func(s, d string) [2]engine.Target {
		return [2]engine.Target{{Name: s, Values: v}, {Name: d, Values: v}}
	}
	switch {
	case v.Bool(EntireInstance):
		return [][2]engine.Target{pair(InstanceName, InstanceName)}, nil
	case v.Bool(config.AllDatabases):
		names, err := src.ListDatabases(ctx, v.List(config.ExcludeDBs))
		if err != nil {
			return nil, fmt.Errorf("error listing databases: %w", err)
		}
		out := make([][2]engine.Target, 0, len(names))
		for _, n := range names {
			out = append(out, pair(n, n))
		}
		return out, nil
	}
	if src.Database == "" {
		return nil, errors.New("database name is required, use DB_NAME environment variable or -d flag")
	}
	if dst.Database == "" {
		return nil, errors.New("target database name is required, use TARGET_DB_NAME environment variable")
	}
	return [][2]engine.Target{pair(src.Database, dst.Database)}, nil
}

// Migrate pipes pg_dump (pg_dumpall for the entire instance) into psql on the
// target. Nothing is written to disk. A database migration runs in one
// transaction on the target; if either side fails, the other is killed
// before psql sees the end of its input.
func (e *Engine) Migrate(ctx context.Context, source, target engine.Target) error {
	v := source.Values
	src, err := SourceConn(v)
	if err != nil {
		return configErr(err)
	}
	dst, err := TargetConn(target.Values)
	if err != nil {
		return configErr(err)
	}
	instance := source.Name == InstanceName
	var (
		dumpTool, srcDB, dstDB string
		dumpArgs               []string
		kind                   = KindPlain
	)
	if instance {
		dumpTool, srcDB, dstDB, kind = "pg_dumpall", src.AuthDB, dst.AuthDB, KindCluster
		dumpArgs = DumpAllArgs(v, src.AuthDB)
	} else {
		dumpTool, srcDB, dstDB = "pg_dump", source.Name, target.Name
		// A migration copies the whole database in plain format,
		// whatever the backup options say.
		dumpArgs = []string{"-w", "-d", ConnInfo(srcDB)}
		created, err := dst.EnsureDatabase(ctx, dstDB)
		if err != nil {
			return err
		}
		if created {
			log().Info("Created the target database", "database", dstDB)
		}
	}
	bin, err := e.dumpBin(ctx, v, src, srcDB, dumpTool)
	if err != nil {
		return err
	}
	tool, args := RestoreCommand(v, kind, dstDB)
	return pipe(ctx,
		execx.Cmd{Name: bin, Args: dumpArgs, Env: src.Env(), Log: log(), StderrLevel: slog.LevelWarn},
		execx.Cmd{Name: e.restoreBin(v, tool), Args: args, Env: dst.Env(), Log: log()},
	)
}

// pipe runs producer | consumer. The first failure closes the pipe with an
// error on both ends, so execx kills the other process.
func pipe(ctx context.Context, producer, consumer execx.Cmd) error {
	pr, pw := io.Pipe()
	producer.Stdout = pw
	consumer.Stdin = pr
	var (
		wg            sync.WaitGroup
		prodErr, cErr error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		prodErr = execx.Run(ctx, producer)
		if prodErr != nil {
			_ = pw.CloseWithError(prodErr)
			return
		}
		_ = pw.Close()
	}()
	go func() {
		defer wg.Done()
		cErr = execx.Run(ctx, consumer)
		if cErr != nil {
			_ = pr.CloseWithError(cErr)
			return
		}
		// Drain so a producer never blocks if the consumer exits early.
		_, _ = io.Copy(io.Discard, pr)
	}()
	wg.Wait()
	if prodErr != nil {
		return prodErr
	}
	return cErr
}
