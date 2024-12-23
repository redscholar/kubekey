//go:build builtin
// +build builtin

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

package builtin

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app"
	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options/builtin"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func newPreCheckCommand() *cobra.Command {
	o := builtin.NewPreCheckOptions()

	cmd := &cobra.Command{
		Use:   "precheck tags...",
		Short: "Check if the nodes is eligible for cluster deployment.",
		Long:  "the tags can specify check items. support: etcd, os, network, cri, nfs.",
		RunE: func(cmd *cobra.Command, args []string) error {
			pipeline, err := o.Complete(cmd, append(args, "playbooks/precheck.yaml"))
			if err != nil {
				return err
			}
			// set workdir
			_const.SetWorkDir(o.WorkDir)
			// create workdir directory,if not exists
			if _, err := os.Stat(o.WorkDir); os.IsNotExist(err) {
				if err := os.MkdirAll(o.WorkDir, os.ModePerm); err != nil {
					return err
				}
			}

			return app.CommandRunE(options.CTX, pipeline, options.Config, options.Inventory)
		},
	}

	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}

func init() {
	app.RegisterInternalCommand(newPreCheckCommand())
}
