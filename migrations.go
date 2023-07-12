package finishline

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"path/filepath"
)

func (f *FinishLine) MigrateUp(dsn string) error {
	m, err := migrate.New("file://"+filepath.ToSlash(f.RootPath)+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("Error running migration: ", err)
		return err
	}

	return nil
}

func (f *FinishLine) MigrateDownAll(dsn string) error {
	m, err := migrate.New("file://"+filepath.ToSlash(f.RootPath)+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil {
		return err
	}

	return nil
}

func (f *FinishLine) Steps(n int, dsn string) error {
	m, err := migrate.New("file://"+filepath.ToSlash(f.RootPath)+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}

	return nil
}

func (f *FinishLine) MigrateForce(dsn string) error {
	m, err := migrate.New("file://"+filepath.ToSlash(f.RootPath)+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}
