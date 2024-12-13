package agent

import (
	"errors"
	"flag"
	"fmt"
	"log"

	bgsvc "github.com/kardianos/service"
	"github.com/kloudmate/km-agent/internal/collector"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"
)

func NewCommand(set otelcol.CollectorSettings) *cobra.Command {
	var col *collector.KmCollector
	var svc bgsvc.Service
	flagSet := flags(featuregate.GlobalRegistry())
	rootCmd := &cobra.Command{
		Use:          set.BuildInfo.Command,
		Version:      set.BuildInfo.Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := updateSettingsUsingFlags(&set, flagSet)
			if err != nil {
				return err
			}

			// col, err := NewCollector(set)
			// if err != nil {
			// 	return err
			// }
			// return col.Run(cmd.Context())
			svc, col, err = NewAgentService(set)
			if err != nil {
				return err
			}
			col.Run(cmd.Context())
			return nil
		},
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install the service",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Attempting to install the service")
			if err := svc.Install(); err != nil {
				fmt.Println(err)
				log.Fatal(err)
			}
			fmt.Println("Service installed successfully")
		},
	}

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := svc.Uninstall(); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Service uninstalled successfully")
		},
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := svc.Start(); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Service started successfully")
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := svc.Stop(); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Service stopped successfully")
		},
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the service directly",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := svc.Logger(nil)
			if err != nil {
				log.Fatal(err)
			}
			if err := svc.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.AddCommand(installCmd, uninstallCmd, startCmd, stopCmd, runCmd)
	rootCmd.AddCommand(newComponentsCommand(set))
	rootCmd.AddCommand(newValidateSubCommand(set, flagSet))
	rootCmd.Flags().AddGoFlagSet(flagSet)
	return rootCmd
}

// Puts command line flags from flags into the CollectorSettings, to be used during config resolution.
func updateSettingsUsingFlags(set *otelcol.CollectorSettings, flags *flag.FlagSet) error {
	resolverSet := &set.ConfigProviderSettings.ResolverSettings
	configFlags := getConfigFlag(flags)

	if len(configFlags) > 0 {
		resolverSet.URIs = configFlags
	}
	if len(resolverSet.URIs) == 0 {
		return errors.New("at least one config flag must be provided")
	}

	if set.ConfigProviderSettings.ResolverSettings.DefaultScheme == "" {
		set.ConfigProviderSettings.ResolverSettings.DefaultScheme = "env"
	}

	if len(resolverSet.ProviderFactories) == 0 {
		return errors.New("at least one Provider must be supplied")
	}
	return nil
}
