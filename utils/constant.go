// Package utils /
package utils

import "os"

const RestoreExample = "restore --dbname database --file db_20231219_022941.sql.gz\n" +
	"restore --dbname database --storage s3 --path /custom-path --file db_20231219_022941.sql.gz"
const BackupExample = "backup --dbname database --disable-compression\n" +
	"backup --dbname database --storage s3 --path /custom-path --disable-compression"

const MainExample = "backup --dbname database --disable-compression\n" +
	"backup --dbname database --storage s3 --path /custom-path\n" +
	"restore --dbname database --file db_20231219_022941.sql.gz"

var Version string

func VERSION(def string) string {
	build := os.Getenv("VERSION")
	if build == "" {
		return def
	}
	return build
}
func FullVersion() string {
	ver := Version
	if b := VERSION(""); b != "" {
		return b
	}
	return ver
}
