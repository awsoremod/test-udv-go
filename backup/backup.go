package backup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"test-udv/pgconn"
	"test-udv/pgpass"
	"time"
)

// Возвращает список файлов, которые находятся в папке backups
func BackupList() ([]fs.DirEntry, error) {
	dir, err := os.Open("./backups")
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// Удаляет файл бэкапа в папке backups
func DeleteBackup(backupFile fs.DirEntry) error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	exPath := filepath.Dir(ex)
	exPath = exPath + `/backups/` + backupFile.Name()

	if err := os.Remove(exPath); err != nil {
		return err
	}
	return nil
}

// Создает бэкап базы данных с помощью утилиты pg_dump. Файлы бэкапов
// помещаются в папку backups
func CreateBackup(entry *pgpass.Entry, database pgconn.Database) error {

	ex, err := os.Executable()
	if err != nil {
		return err
	}

	exPath := filepath.Dir(ex)

	t := time.Now()
	year := t.Year()
	month := int(t.Month())
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()

	str := fmt.Sprintf("_%d-%d-%d_%d-%d", year, month, day, hour, minute)
	queries := []string{
		`/C`,
		`pg_dump.exe`,
		`-h` + entry.Host,
		`-p` + entry.Port,
		`-U` + entry.User,
		`--no-password`,
		`-f` + exPath + `/backups/` + database.Name + str + `.dump`,
		`-F` + `c`,
		`-d` + database.Name,
	}

	cmd := exec.Command("cmd", queries...)
	cmd.Dir = exPath + "/pg_dump_restore_15_2"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprint(err) + ": " + string(output))
	}
	return nil
}

// Восстанавливает базу данных из бэкапа с помощью утилиты pg_restore
func BackupRestore(entry *pgpass.Entry, backupFile fs.DirEntry) error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	exPath := filepath.Dir(ex)

	queries := []string{
		`/C`,
		`pg_restore.exe`,
		`-h` + entry.Host,
		`-p` + entry.Port,
		`-U` + entry.User,
		`--no-password`,
		`-F` + `c`,
		`-C`, // TODO проверить с базами
		//`-c`,
		//`--if-exists`,
		//`-f` + `-`,
		`-d` + entry.Dbname,
		exPath + `/backups/` + backupFile.Name(),
	}

	cmd := exec.Command("cmd", queries...)
	cmd.Dir = exPath + "/pg_dump_restore_15_2"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprint(err) + ": " + string(output[:500]))
	}
	return nil
}
