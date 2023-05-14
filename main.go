package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"test-udv/backup"
	"test-udv/pgconn"
	"test-udv/pgpassfile"
)

const (
	databaseList    = 1
	backupList      = 2
	deleteDatabase  = 3
	deleteBackup    = 4
	createBackup    = 5
	restoreDatabase = 6
	exit            = 7
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
		//os.Exit(1)
		return
	}

	pgpass, err := pgpassfile.NewPgpass()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}

	pgpassFile, err := pgpass.CreateOrOpenFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}
	defer pgpassFile.Close()

	var entry *pgpassfile.Entry
	if len(pgpass.Entries) == 0 {
		entry = &pgpassfile.Entry{
			Host:     "localhost",
			Port:     "5432",
			Dbname:   "",
			User:     "",
			Password: "",
		}
	} else {
		copyLastEntry := *pgpass.Entries[len(pgpass.Entries)-1]
		entry = &copyLastEntry
	}

	if err := enterСonfigParameters(entry); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}

	ctx := context.Background()
	conn, err := pgconn.OpenConnect(ctx, entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}
	defer conn.Close(ctx)

	if err := pgconn.GetVersion(ctx, conn); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}

	isExistConfig, err := pgpass.IsExistEntry(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}
	if !isExistConfig {
		str := "Добавить данное подключение в файл конфигурации?\n" +
			"1: Да\n" +
			"2: Нет\n"
		fmt.Print(str)
		fmt.Print(`Ваш ответ (Нет): `)
		var selectedNumber int
		if _, err := fmt.Scanln(&selectedNumber); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			//os.Exit(1)
			return
		}
		if selectedNumber == 1 {
			if err := pgpass.AddConfigInFile(pgpassFile, entry); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				//os.Exit(1)
				return
			}
		}
	}
	pgpassFile.Close()

	if err := os.Rename(pgpass.Path, pgpass.Path+`_origin`); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}
	tmpPgpassFile, err := pgpass.CreateOrOpenFile()
	if err != nil {
		tmpPgpassFile.Close()
		if err := os.Rename(pgpass.Path+`_origin`, pgpass.Path); err != nil {
			fmt.Fprintf(os.Stderr, "Не получилось обратно переименовать оригинальный файл, переименуйте файл pgpass.conf_origin в pgpass.conf %v\n", err)
			//os.Exit(1)
			return
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		//os.Exit(1)
		return
	}
	defer deleteTmpFile(pgpass.Path)
	defer tmpPgpassFile.Close()

	// Setup our Ctrl+C handler
	setupCloseHandler(pgpass.Path)

	if err := pgpass.AddConfigInFile(tmpPgpassFile, &pgpassfile.Entry{Host: "*", Port: "*", Dbname: "*", User: "*", Password: entry.Password}); err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось записать во временный файл конфигурации. %v\n", err)
		//os.Exit(1)
		return
	}
	tmpPgpassFile.Close()

menuСycle:
	for {
		str := "Меню\n" +
			"1: Список баз данных\n" +
			"2: Список бэкапов\n" +
			"3: Удалить базу данных\n" +
			"4: Удалить бэкап\n" +
			"5: Создать бэкап базы данных\n" +
			"6: Восстановить базу данных из бэкапа\n" +
			"7: Выйти\n\n"
		fmt.Print(str)
		fmt.Print("Choose a option: ")
		var inputOption int
		if _, err := fmt.Scanln(&inputOption); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		switch inputOption {
		case databaseList:
			databases, err := pgconn.DatabaseList(ctx, conn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println(`Вот список всех баз данных:`)
			for _, d := range databases {
				fmt.Printf("%s\n", d.Name) // вынести в функцию
			}

		case backupList:
			files, err := backup.BackupList()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println(`Вот список всех бэкапов:`)
			for _, file := range files {
				fmt.Println(file.Name()) // вынести в функцию
			}

		case deleteDatabase:
			databases, err := pgconn.DatabaseList(ctx, conn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println(`Вот список всех баз данных:`)
			for i, d := range databases {
				fmt.Printf("%d: %s\n", i, d.Name)
			}

			fmt.Print(`Выбирите базу данных для удаления: `)
			var indexDatabase int
			if _, err := fmt.Scanln(&indexDatabase); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			// ошибка если введено число больше размера databases
			str := "Вы уверены, что хотите удалить базу данных " + databases[indexDatabase].Name + "?\n" +
				"1: Да\n" +
				"2: Нет, отмена\n"
			fmt.Print(str)
			fmt.Print(`Введите ваш ответ: `)
			var selectedNumber int
			if _, err := fmt.Scanln(&selectedNumber); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			switch selectedNumber {
			case 1:
				if err = pgconn.DeleteDatabase(ctx, conn, databases[indexDatabase]); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					continue
				}
				fmt.Printf("База данных %s успешно удалилась.\n", databases[indexDatabase].Name)

			default:
				continue
			}

		case deleteBackup:
			backups, err := backup.BackupList()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println(`Вот список всех бэкапов:`)
			for i, file := range backups {
				fmt.Printf("%d: %s\n", i, file.Name())
			}
			fmt.Print(`Выбирите бэкап для удаления: `)
			var indexBackup int
			if _, err := fmt.Scanln(&indexBackup); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			// ошибка если введено число больше размера backups
			str := "Вы уверены, что хотите удалить бэкап " + backups[indexBackup].Name() + " :\n" +
				"1: Да\n" +
				"2: Нет, отмена\n"
			fmt.Print(str)
			fmt.Print(`Введите ваш ответ: `)
			var selectedNumber int
			if _, err := fmt.Scanln(&selectedNumber); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			switch selectedNumber {
			case 1:
				if err = backup.DeleteBackup(backups[indexBackup]); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					continue
				}
				fmt.Printf("Бэкап %s успешно удалён.\n", backups[indexBackup].Name())

			default:
				continue
			}

		case createBackup:
			databases, err := pgconn.DatabaseList(ctx, conn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println(`Вот список всех баз данных:`)
			for i, d := range databases {
				fmt.Printf("%d: %s\n", i, d.Name)
			}

			fmt.Print(`Выбирите базу данных для создания бэкапа: `)
			var indexDatabase int
			if _, err := fmt.Scanln(&indexDatabase); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			// ошибка если введено число больше размера databases
			if err := backup.CreateBackup(entry, databases[indexDatabase]); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println("Успешно создан бэкап базы данных " + databases[indexDatabase].Name)

		case restoreDatabase:
			backups, err := backup.BackupList()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println(`Вот список всех бэкапов:`)
			for i, file := range backups {
				fmt.Printf("%d: %s\n", i, file.Name())
			}
			fmt.Print(`Выбирите бэкап для востановления бд: `)
			var indexBackup int
			if _, err := fmt.Scanln(&indexBackup); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			// ошибка если введено число больше размера backups
			if err := backup.BackupRestore(entry, backups[indexBackup]); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Println("Успешно восстановлена база данных из файла " + backups[indexBackup].Name())

		case exit:
			break menuСycle
		default:
			fmt.Println(`Ошибка ввода.`)
			continue
		}

		fmt.Print(`Нажмите ENTER для перехода в меню: `)
		var input int
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		fmt.Println() // если ввести много пробелов то много раз попадаю сюда
	}
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

func deleteTmpFile(pathPgpass string) error {

	if err := os.Remove(pathPgpass); err != nil {
		fmt.Fprintf(os.Stderr, "Error when deleting a file: %v\n", err)
		return err
	}

	fileName := pathPgpass + `_origin`
	if err := os.Rename(fileName, pathPgpass); err != nil {
		fmt.Fprintf(os.Stderr, "Error when trying to rename file '%s': %v\n", fileName, err)
		return err
	}

	return nil
}

func enterСonfigParameters(entry *pgpassfile.Entry) error {
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
			fmt.Fprintf(os.Stderr, "Error when entering configuration parameters %v\n", err)
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
