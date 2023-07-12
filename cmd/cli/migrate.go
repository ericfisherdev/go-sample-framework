package main

func doMigrate(arg2, arg3 string) error {
	dsn := getDSN()

	// run the migration command
	switch arg2 {
	case "up":
		err := fin.MigrateUp(dsn)
		if err != nil {
			return err
		}
	case "down":
		if arg3 == "all" {
			err := fin.MigrateDownAll(dsn)
			if err != nil {
				return err
			}
		} else {
			err := fin.Steps(-1, dsn)
			if err != nil {
				return err
			}
		}
	case "reset":
		err := fin.MigrateDownAll(dsn)
		if err != nil {
			return err
		}
		err = fin.MigrateUp(dsn)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}
	return nil
}
