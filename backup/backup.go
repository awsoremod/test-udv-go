package backup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"test-udv/pgconn"
	"test-udv/pgpassfile"
	"time"
)

func BackupList() ([]fs.DirEntry, error) {
	// TODO сделать возврат только файлов .dump, .tar
	// Переделать на получение открытого файла

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

func CreateBackup(entry *pgpassfile.Entry, database pgconn.Database) error {
	// TODO добавить 0 перед месяцем если меньше 10
	// Проверить использование форматов времени
	// Проверить на ошибку если в имени базы данных есть пробел
	// Возможно имена папок вынести в константы

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

func BackupRestore(entry *pgpassfile.Entry, backupFile fs.DirEntry) error {
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
		return errors.New(fmt.Sprint(err) + ": " + string(output[:300]))
	}
	return nil
}
