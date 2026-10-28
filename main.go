// Command pg-bkup backs up, restores and migrates PostgreSQL databases.
package main

import (
	"github.com/jkaninda/bkupkit/cli"
	"github.com/jkaninda/pg-bkup/internal/pg"
)

var Version = "dev"

func main() {
	app := &cli.App{
		Engine:  pg.New(),
		Version: Version,
		Short:   "Back up, restore and migrate PostgreSQL databases (local, S3, GCS, Azure, SFTP, FTP)",
		Extend:  pg.Extend,
	}
	app.Main()
}
