package agent

import (
	"flag"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/collector/otelcol"
)

// newValidateSubCommand constructs a new validate sub command using the given CollectorSettings.
func newValidateSubCommand(set otelcol.CollectorSettings, flagSet *flag.FlagSet) *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validates the config without running the collector",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := updateSettingsUsingFlags(&set, flagSet); err != nil {
				return err
			}
			col, err := otelcol.NewCollector(set)
			if err != nil {
				return err
			}
			return col.DryRun(cmd.Context())
		},
	}
	validateCmd.Flags().AddGoFlagSet(flagSet)
	return validateCmd
}
