package main

import (
	"context"
	"errors"
	"fmt"
	"test-udv/backup"
	"test-udv/pgconn"
	"test-udv/pgpass"

	"github.com/jackc/pgx/v5"
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

func menuCycle(isExit *bool, conn *pgx.Conn, entry *pgpass.Entry) error {
	ctx := context.Background()
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
		return err
	}

	switch inputOption {
	case databaseList:
		databases, err := pgconn.DatabaseList(ctx, conn)
		if err != nil {
			return err
		}
		fmt.Println(`Вот список всех баз данных:`)
		for _, d := range databases {
			fmt.Printf("%s\n", d.Name) // вынести в функцию
		}

	case backupList:
		files, err := backup.BackupList()
		if err != nil {
			return err
		}
		fmt.Println(`Вот список всех бэкапов:`)
		for _, file := range files {
			fmt.Println(file.Name()) // вынести в функцию
		}

	case deleteDatabase:
		databases, err := pgconn.DatabaseList(ctx, conn)
		if err != nil {
			return err
		}
		fmt.Println(`Вот список всех баз данных:`)
		for i, d := range databases {
			fmt.Printf("%d: %s\n", i, d.Name)
		}

		fmt.Print(`Выбирите базу данных для удаления: `)
		var indexDatabase int
		if _, err := fmt.Scanln(&indexDatabase); err != nil {
			return err
		}
		// ошибка если введено число больше размера databases
		str := "Вы уверены, что хотите удалить базу данных " + databases[indexDatabase].Name + "?\n" +
			"1: Да\n" +
			"2: Нет, отмена\n"
		fmt.Print(str)
		fmt.Print(`Введите ваш ответ: `)
		var selectedNumber int
		if _, err := fmt.Scanln(&selectedNumber); err != nil {
			return err
		}
		switch selectedNumber {
		case 1:
			if err = pgconn.DeleteDatabase(ctx, conn, databases[indexDatabase]); err != nil {
				return err
			}
			fmt.Printf("База данных %s успешно удалилась.\n", databases[indexDatabase].Name)

		default:
			return nil
		}

	case deleteBackup:
		backups, err := backup.BackupList()
		if err != nil {
			return err
		}
		fmt.Println(`Вот список всех бэкапов:`)
		for i, file := range backups {
			fmt.Printf("%d: %s\n", i, file.Name())
		}
		fmt.Print(`Выбирите бэкап для удаления: `)
		var indexBackup int
		if _, err := fmt.Scanln(&indexBackup); err != nil {
			return err
		}
		// ошибка если введено число больше размера backups
		str := "Вы уверены, что хотите удалить бэкап " + backups[indexBackup].Name() + " :\n" +
			"1: Да\n" +
			"2: Нет, отмена\n"
		fmt.Print(str)
		fmt.Print(`Введите ваш ответ: `)
		var selectedNumber int
		if _, err := fmt.Scanln(&selectedNumber); err != nil {
			return err
		}
		switch selectedNumber {
		case 1:
			if err = backup.DeleteBackup(backups[indexBackup]); err != nil {
				return err
			}
			fmt.Printf("Бэкап %s успешно удалён.\n", backups[indexBackup].Name())

		default:
			return nil
		}

	case createBackup:
		databases, err := pgconn.DatabaseList(ctx, conn)
		if err != nil {
			return err
		}
		fmt.Println(`Вот список всех баз данных:`)
		for i, d := range databases {
			fmt.Printf("%d: %s\n", i, d.Name)
		}

		fmt.Print(`Выбирите базу данных для создания бэкапа: `)
		var indexDatabase int
		if _, err := fmt.Scanln(&indexDatabase); err != nil {
			return err
		}
		// ошибка если введено число больше размера databases
		if err := backup.CreateBackup(entry, databases[indexDatabase]); err != nil {
			return err
		}
		fmt.Println("Успешно создан бэкап базы данных " + databases[indexDatabase].Name)

	case restoreDatabase:
		backups, err := backup.BackupList()
		if err != nil {
			return err
		}
		fmt.Println(`Вот список всех бэкапов:`)
		for i, file := range backups {
			fmt.Printf("%d: %s\n", i, file.Name())
		}
		fmt.Print(`Выбирите бэкап для востановления бд: `)
		var indexBackup int
		if _, err := fmt.Scanln(&indexBackup); err != nil {
			return err
		}
		// ошибка если введено число больше размера backups
		if err := backup.BackupRestore(entry, backups[indexBackup]); err != nil {
			return err
		}
		fmt.Println("Успешно восстановлена база данных из файла " + backups[indexBackup].Name())

	case exit:
		*isExit = true
		return nil
	default:
		return errors.New(`ошибка ввода`)
	}

	return nil
}
