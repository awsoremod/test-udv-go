package pgpass

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Entry represents a line in a PG passfile.
type Entry struct {
	Host     string
	Port     string
	Dbname   string
	User     string
	Password string
}

func (e *Entry) string() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", e.Host, e.Port, e.Dbname, e.User, e.Password)
}

// parseLine parses a line into an *Entry.
func parseLine(line string) (*Entry, error) {
	const (
		tmpBackslash = "\r"
		tmpColon     = "\n"
	)

	line = strings.TrimSpace(line)

	if strings.HasPrefix(line, "#") {
		return nil, errors.New("начинается на комментарий")
	}

	line = strings.Replace(line, `\\`, tmpBackslash, -1)
	line = strings.Replace(line, `\:`, tmpColon, -1)

	parts := strings.Split(line, ":")
	if len(parts) != 5 {
		return nil, errors.New("неправильное количество двоеточий")
	}

	// Unescape escaped colons and backslashes
	for i := range parts {
		parts[i] = strings.Replace(parts[i], tmpBackslash, `\`, -1)
		parts[i] = strings.Replace(parts[i], tmpColon, `:`, -1)
	}

	return &Entry{
		Host:     parts[0],
		Port:     parts[1],
		Dbname:   parts[2],
		User:     parts[3],
		Password: parts[4],
	}, nil
}

func GetPath() (string, error) {
	// TODO : Добавить возврат в зависимости от операционной системы

	pathPgpass, ok := os.LookupEnv(`APPDATA`)
	if !ok {
		return "", errors.New("not in environment variables APPDATA")
	}

	pathPgpass = pathPgpass + `\postgresql\pgpass.conf`
	return pathPgpass, nil
}

// Читает файл по пути и парсит его в массив подключений
func GetEntries(path string) ([]*Entry, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries, err := parsePassfile(f)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// Парсит файл в массив подключений
func parsePassfile(r io.Reader) ([]*Entry, error) {
	// Если функция вызывается повторно на том же файле,
	// то сканирование будет происходить с прошлого места.

	entries := make([]*Entry, 0, 10)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		entry, err := parseLine(scanner.Text())
		if err == nil {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// Проверяет существование файла Pgpass
func IsExistFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Если файл существут, то функция откроет файл
func CreateOrOpenFile(path string) (*os.File, error) {
	isExist, err := IsExistFile(path)
	if err != nil {
		return nil, err
	}
	if isExist {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0600); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Проверяет, есть ли строка подключения в виде *Entry
// в файле
func IsExistEntry(path string, entry *Entry) (bool, error) {
	// TODO : Проверять entry на nil

	entries, err := GetEntries(path)
	if err != nil {
		return false, err
	}

	for _, e := range entries {
		// TODO : Вынести в функцию
		if (e.Host == entry.Host) &&
			(e.Port == entry.Port) &&
			(e.Dbname == entry.Dbname) &&
			(e.User == entry.User) &&
			(e.Password == entry.Password) {
			return true, nil
		}
	}
	return false, nil
}

// Добавляет строку соединения в виде *Entry
// в файл конфигурации
func AddEntryInFile(file *os.File, config *Entry) error {
	// TODO : Проверять на nil

	configString := config.string()

	str := fmt.Sprintf("\n%s\n", configString)

	if _, err := file.WriteString(str); err != nil {
		return err
	}

	return nil
}
