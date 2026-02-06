package main

import "github.com/spf13/cobra"

func NewRBLNValidatorApp() *cobra.Command {
	builder := newConfigBuilder()

	cmd := &cobra.Command{
		Use:           "rbln-validator",
		Short:         "RBLN NPU operator validator.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newDriverCommand(builder),
		newToolkitCommand(builder),
	)

	builder.bindFlags(cmd)

	return cmd
}
