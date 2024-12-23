/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"fmt"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

var internalCommand = make([]*cobra.Command, 0)

func RegisterInternalCommand(command *cobra.Command) {
	for _, c := range internalCommand {
		if c.Name() == command.Name() {
			// command has register. skip
			return
		}
	}
	internalCommand = append(internalCommand, command)
}

// NewRootCommand console command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "kk",
		Long: "kubekey is a daemon that execute command in a node",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := options.InitGOPS(); err != nil {
				return err
			}

			return options.InitProfiling(options.CTX)
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return options.FlushProfiling()
		},
	}
	cmd.SetContext(options.CTX)

	// add common flag
	flags := cmd.PersistentFlags()
	options.AddProfilingFlags(flags)
	options.AddKlogFlags(flags)
	options.AddGOPSFlags(flags)

	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newPipelineCommand())
	cmd.AddCommand(newVersionCommand())
	// internal command
	cmd.AddCommand(internalCommand...)

	return cmd
}

// CommandRunE. run pipeline use command and return a error.
func CommandRunE(ctx context.Context, pipeline *kkcorev1.Pipeline, config *kkcorev1.Config, inventory *kkcorev1.Inventory) error {
	restconfig, err := proxy.NewConfig(&rest.Config{})
	if err != nil {
		return fmt.Errorf("could not get rest config: %w", err)
	}
	client, err := ctrlclient.New(restconfig, ctrlclient.Options{
		Scheme: _const.Scheme,
	})
	if err != nil {
		return fmt.Errorf("could not get runtime-client: %w", err)
	}
	// create inventory
	if err := client.Create(ctx, inventory); err != nil {
		klog.ErrorS(err, "Create inventory error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))

		return err
	}
	// create pipeline
	// pipeline.Status.Phase = kkcorev1.PipelinePhaseRunning
	if err := client.Create(ctx, pipeline); err != nil {
		klog.ErrorS(err, "Create pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))

		return err
	}

	return manager.NewCommandManager(manager.CommandManagerOptions{
		Pipeline:  pipeline,
		Config:    config,
		Inventory: inventory,
		Client:    client,
	}).Run(ctx)
}
