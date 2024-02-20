package utils

const RestoreExample = "pg-bkup restore --dbname database --file db_20231219_022941.sql.gz\n" +
	"bkup restore --dbname database --storage s3 --path /custom-path --file db_20231219_022941.sql.gz"
const BackupExample = "pg-bkup backup --dbname database --disable-compression\n" +
	"pg-bkup backup --dbname database --storage s3 --path /custom-path --disable-compression"

const MainExample = "pg-bkup backup --dbname database --disable-compression\n" +
	"pg-bkup backup --dbname database --storage s3 --path /custom-path\n" +
	"pg-bkup restore --dbname database --file db_20231219_022941.sql.gz"
