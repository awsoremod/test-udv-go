package pgpassfile

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

type Pgpass struct {
	// TODO метод open добавить
	// Добавить просмотр env postgres параметров

	Path    string
	Entries []*Entry
}

func NewPgpass() (*Pgpass, error) {
	path, err := getPath()
	if err != nil {
		return nil, err
	}

	entryes, err := readPassfile(path)
	if err != nil {
		return nil, err
	}
	return &Pgpass{path, entryes}, nil
}

func getPath() (string, error) {
	// TODO : Добавить, изменить возврат в зависимости от операционной системы

	pathPgpass, ok := os.LookupEnv(`APPDATA`)
	if !ok {
		return "", errors.New("not in environment variables APPDATA")
	}

	pathPgpass = pathPgpass + `\postgresql\pgpass.conf`
	return pathPgpass, nil
}

// ReadPassfile reads the file at path and parses it into a Passfile.
func readPassfile(path string) ([]*Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entryes, err := parsePassfile(f)
	if err != nil {
		return nil, err
	}
	return entryes, nil
}

// ParsePassfile reads r and parses it into a Passfile.
func parsePassfile(r io.Reader) ([]*Entry, error) {
	// Если функция вызывается повторно на том же файле,
	// то сканирование будет происходить с прошлого места. TODO перепроверить

	entryes := make([]*Entry, 0, 10)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		entry, err := parseLine(scanner.Text())
		if err == nil {
			entryes = append(entryes, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entryes, nil
}

// Проверяет существования файла Pgpass
func (p *Pgpass) IsExistFile() (bool, error) {
	_, err := os.Stat(p.Path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Если файл существут, то функция откроет файл
func (p *Pgpass) CreateOrOpenFile() (*os.File, error) {
	isExist, err := p.IsExistFile()
	if err != nil {
		return nil, err
	}
	if isExist {
		file, err := os.OpenFile(p.Path, os.O_RDWR|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	if err := os.MkdirAll(filepath.Dir(p.Path), 0600); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(p.Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (p *Pgpass) IsExistEntry(entry *Entry) (bool, error) {
	// TODO
	// Проверять config на nil
	// Проверить p.Entries на nil
	// Перегрузить оператор сравнения двух структур, загуглить

	for _, e := range p.Entries {
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

// Добавляет строку соединения в конфиг
func (p *Pgpass) AddConfigInFile(file *os.File, config *Entry) error {
	// Проверять config на nil
	// Возможно изменить file на какой-нибудь интерфейс
	// Проверить как работает WriteString
	p.Entries = append(p.Entries, config)

	configString := config.string()

	if _, err := file.WriteString(configString); err != nil {
		return err
	}

	return nil
}
