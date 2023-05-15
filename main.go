package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"test-udv/pgconn"
	"test-udv/pgpass"

	"github.com/jackc/pgx/v5"
)

func main() {
	// TODO : Добавить --help
	// Вынести что-нибудь в параметры утилиты
	// Предложить пользователю выбрать строку подключения из строк подключения в файле конфигурации pgpass.conf

	defer func() {
		var enterToClose int
		fmt.Scanln(&enterToClose)
	}()

	if err := os.MkdirAll("backups", 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Create dir 'backups' error: %v\n", err)
		return
	}

	pgpassFile, err := pgpassCreateOrOpen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer pgpassFile.Close()

	pathToPgpass := pgpassFile.Name()

	entry, err := configureEntry(pathToPgpass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	ctx := context.Background()
	conn, err := pgconn.OpenConnect(ctx, entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer conn.Close(ctx)

	if err := offerToSaveEntry(pgpassFile, entry); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	pgpassFile.Close()

	if err := createTmpPgpassFile(pathToPgpass, entry); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer backOrigin(pathToPgpass)

	setupCloseHandler(pathToPgpass)

	menuLoop(ctx, conn, entry)
}

// Возвращает открытый или созданный файл pgpass.conf
func pgpassCreateOrOpen() (*os.File, error) {
	path, err := pgpass.GetPath()
	if err != nil {
		return nil, err
	}

	file, err := pgpass.CreateOrOpenFile(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Конфигурирует строку подключения
func configureEntry(path string) (*pgpass.Entry, error) {
	entries, err := pgpass.GetEntries(path)
	if err != nil {
		return nil, err
	}

	entry, err := buildDefaultEntry(entries)
	if err != nil {
		return nil, err
	}

	if err := enterСonfigParameters(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// Если в файле конфигурации есть строки подключения, возвращает
// последнюю строку подключения. Если нет, формирует строку
// по умолчанию
func buildDefaultEntry(entries []*pgpass.Entry) (*pgpass.Entry, error) {
	if len(entries) == 0 {
		entry := &pgpass.Entry{
			Host:     "localhost",
			Port:     "5432",
			Dbname:   "",
			User:     "",
			Password: "",
		}
		return entry, nil
	}

	copyLastEntry := *entries[len(entries)-1]
	entry := &copyLastEntry

	return entry, nil
}

// Ввод параметров подключения
func enterСonfigParameters(entry *pgpass.Entry) error {
	// TODO убрать отображение вводимого и предлагаемого пароля

	if err := enterСonfigParameter("Host", &entry.Host); err != nil {
		return err
	}
	if err := enterСonfigParameter("Port", &entry.Port); err != nil {
		return err
	}
	if err := enterСonfigParameter("Dbname", &entry.Dbname); err != nil {
		return err
	}
	if err := enterСonfigParameter("User", &entry.User); err != nil {
		return err
	}
	if err := enterСonfigParameter("Password", &entry.Password); err != nil {
		return err
	}
	return nil
}

// Ввод параметра подключения
func enterСonfigParameter(str string, variable *string) error {
	// TODO : Добавить условие, если в параметре ничего нет,
	// нельзя нажать enter.
	// Добавить проверки входных параметров.

	fmt.Printf("Укажите %s (%s): ", str, *variable)
	var input string
	n, err := fmt.Scanln(&input)
	isInputClear := n == 0
	if err != nil {
		if !isInputClear {
			return err
		}
	}
	if !isInputClear {
		input = strings.TrimSpace(input)
		*variable = input
		return nil
	}

	return nil
}

// Если данное подключение существует в файле конфигурации
// то функция ничего не предлагает
func offerToSaveEntry(pgpassFile *os.File, entry *pgpass.Entry) error {
	isExistConfig, err := pgpass.IsExistEntry(pgpassFile.Name(), entry)
	if err != nil {
		return err
	}
	if !isExistConfig {
		str := "Добавить данное подключение в файл конфигурации?\n" +
			"1: Да\n" +
			"2: Нет\n"
		fmt.Print(str)
		fmt.Print(`Ваш ответ (Нет): `)
		var selectedNumber int
		if _, err := fmt.Scanln(&selectedNumber); err != nil {
			return err
		}
		if selectedNumber == 1 {
			if err := pgpass.AddEntryInFile(pgpassFile, entry); err != nil {
				return err
			}
		}
	}
	return nil
}

// Создает новый временный файл pgpass.conf. Старый файл переименуется
// в pgpass.conf_origin
func createTmpPgpassFile(path string, entry *pgpass.Entry) error {
	if err := os.Rename(path, path+`_origin`); err != nil {
		return err
	}

	tmpPgpassFile, err := pgpass.CreateOrOpenFile(path)
	if err != nil {
		if err := os.Rename(path+`_origin`, path); err != nil {
			str := fmt.Sprintf("Не получилось обратно переименовать оригинальный файл,"+
				"переименуйте файл pgpass.conf_origin в pgpass.conf. error: %v\n", err)
			return errors.New(str) // TODO : добавить информацию о верхней ошибке
		}
		return err
	}
	defer tmpPgpassFile.Close()

	superEntry := &pgpass.Entry{Host: "*", Port: "*", Dbname: "*", User: "*", Password: entry.Password}
	if err := pgpass.AddEntryInFile(tmpPgpassFile, superEntry); err != nil {
		if err := deleteTmpFile(path); err != nil {
			return err // TODO : добавить информацию о верхней ошибке
		}
		str := fmt.Sprintf("Не удалось записать во временный файл конфигурации. %v\n", err)
		return errors.New(str)
	}
	return nil
}

// Setup our Ctrl+C handler
func setupCloseHandler(pathPgpass string) {
	// TODO : Добавить корректное закрытие конекта к базе

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		if err := deleteTmpFile(pathPgpass); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("- Good bye!")
		os.Exit(0)
	}()
}

// Удаляет временный файл и переименовывает pgpass.conf_origin в pgpass.conf.
// Выводит сообщения об ошибках в os.Stderr
func backOrigin(pathPgpass string) {
	if err := deleteTmpFile(pathPgpass); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

// Удаляет временный файл и переименовывает pgpass.conf_origin в pgpass.conf.
func deleteTmpFile(pathPgpass string) error {
	if err := os.Remove(pathPgpass); err != nil {
		return err
	}

	fileName := pathPgpass + `_origin`
	if err := os.Rename(fileName, pathPgpass); err != nil {
		return err
	}

	return nil
}

// Цикл отображения меню
func menuLoop(ctx context.Context, conn *pgx.Conn, entry *pgpass.Entry) {
	for {
		fmt.Println()

		isExit, err := menu(ctx, conn, entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		if isExit {
			break
		}

		fmt.Print(`Нажмите ENTER`)
		var input int
		fmt.Scanln(&input)
	}
}
