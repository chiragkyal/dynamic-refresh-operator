package main

import (
	"os"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"

	"github.com/chiragkyal/dynamic-refresh-operator/pkg/operator"
	"github.com/chiragkyal/dynamic-refresh-operator/pkg/version"
)

func main() {
	command := NewOperatorCommand()
	code := cli.Run(command)
	os.Exit(code)
}

func NewOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dynamic-refresh-operator",
		Short: "Dynamic Refresh Operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	ctrlCmd := controllercmd.NewControllerCommandConfig(
		"dynamic-refresh-operator",
		version.Get(),
		operator.RunOperator,
	).NewCommand()
	ctrlCmd.Use = "start"
	ctrlCmd.Short = "Start the Dynamic Refresh Operator"

	cmd.AddCommand(ctrlCmd)

	return cmd
}
