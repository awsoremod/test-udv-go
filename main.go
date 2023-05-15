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
)

func main() {
	// TODO Добавить --help
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

	pathToPgpass := pgpassFile.Name() // добавить const

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

	// Setup our Ctrl+C handler
	setupCloseHandler(pathToPgpass)

	var isExit bool
	for {
		fmt.Println()
		err := menuCycle(&isExit, conn, entry) // добавить контекст
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		if isExit {
			break
		}

		fmt.Print(`Нажмите ENTER`)
		var input int
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

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

func enterСonfigParameter(str string, variable *string) error {
	// TODO Eсли пустота уже есть в конфиге то нельзя нажать enter
	// Добавить проверки входных параметров

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

func offerToSaveEntry(pgpassFile *os.File, entry *pgpass.Entry) error {
	isExistConfig, err := pgpass.IsExistEntry(pgpassFile.Name(), entry) // переделать
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
			if err := pgpass.AddConfigInFile(pgpassFile, entry); err != nil {
				return err
			}
		}
	}
	return nil
}

func createTmpPgpassFile(path string, entry *pgpass.Entry) error {
	if err := os.Rename(path, path+`_origin`); err != nil {
		return err
	}

	tmpPgpassFile, err := pgpass.CreateOrOpenFile(path)
	if err != nil {
		if err := os.Rename(path+`_origin`, path); err != nil {
			str := fmt.Sprintf("Не получилось обратно переименовать оригинальный файл, переименуйте файл pgpass.conf_origin в pgpass.conf. error: %v\n", err)
			return errors.New(str) // добавить информацию о верхней ошибке
		}
		return err
	}
	defer tmpPgpassFile.Close()

	if err := pgpass.AddConfigInFile(tmpPgpassFile, &pgpass.Entry{Host: "*", Port: "*", Dbname: "*", User: "*", Password: entry.Password}); err != nil {
		if err := deleteTmpFile(path); err != nil {
			return err // добавить информацию о верхней ошибке
		}
		str := fmt.Sprintf("Не удалось записать во временный файл конфигурации. %v\n", err)
		return errors.New(str)
	}
	return nil
}

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

func backOrigin(pathPgpass string) {
	if err := deleteTmpFile(pathPgpass); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

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
