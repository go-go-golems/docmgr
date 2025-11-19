package workspace

import "github.com/spf13/cobra"

// Attach registers workspace-wide commands (init/configure/status/doctor).
func Attach(root *cobra.Command) error {
	initCmd, err := newInitCommand()
	if err != nil {
		return err
	}
	configureCmd, err := newConfigureCommand()
	if err != nil {
		return err
	}
	statusCmd, err := newStatusCommand()
	if err != nil {
		return err
	}
	doctorCmd, err := newDoctorCommand()
	if err != nil {
		return err
	}

	root.AddCommand(initCmd, configureCmd, statusCmd, doctorCmd)
	return nil
}
