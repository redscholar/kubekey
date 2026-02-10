package app

import (
	"flag"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/kubesphere/kubekey/v4/cmd/apiserver/app/options"
	"github.com/kubesphere/kubekey/v4/pkg/apiserver"
)

// ApiServerCommand creates a new cobra command for starting the KubeKey api server
// It initializes the web server with the provided configuration options and starts
// the HTTP server with web UI interface
func NewAPIServerCommand() *cobra.Command {
	o := options.NewAPIServerOptions()
	ctx := signals.SetupSignalHandler()

	cmd := &cobra.Command{
		Use:   "apiserver",
		Short: "start a http api server for web installer.",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := options.InitGOPS(); err != nil {
				return err
			}

			return options.InitProfiling(ctx)
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return options.FlushProfiling()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			svr := apiserver.NewAPIServer(o)

			// Initialize and run the web manager with provided options
			return svr.Run(ctx)
		},
	}
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.ReplaceAll(fl.Name, "_", "-")
		cmd.Flags().AddGoFlag(fl)
	})
	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}

	return cmd
}
