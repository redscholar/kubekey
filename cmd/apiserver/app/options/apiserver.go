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

package options

import (
	"fmt"
	"os"
	"path/filepath"

	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

// ControllerManagerServerOptions for NewControllerManagerServerOptions
type APIServerOptions struct {
	Local                   bool
	Port                    int
	MaxConcurrentReconciles int

	Workdir    string
	UIPath     string
	SchemaPath string
}

// NewControllerManagerServerOptions for NewControllerManagerCommand
func NewAPIServerOptions() *APIServerOptions {
	// Set the working directory to the current directory joined with "kubekey".
	wd, err := os.Getwd()
	if err != nil {
		klog.Warningf("get current dir error: %v, use home dir instead", err)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			klog.Warningf("get home dir error: %v, use / instead", err)
			wd = "/"
		} else {
			wd = homeDir
		}
	}

	return &APIServerOptions{
		Local:                   true,
		Port:                    8080,
		MaxConcurrentReconciles: 1,
		Workdir:                 filepath.Join(wd, "kubekey"),
		UIPath:                  filepath.Join(wd, "dist"),
		SchemaPath:              filepath.Join(wd, "schema"),
	}
}

// Flags add to NewControllerManagerCommand
func (o *APIServerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	cfs := fss.FlagSet("apiserver")
	cfs.BoolVar(&o.Local, "port", o.Local, fmt.Sprintf("The number of port for apiserver.default is %v", o.Local))
	cfs.IntVar(&o.Port, "port", o.Port, fmt.Sprintf("The number of port for apiserver.default is %v", o.Port))
	cfs.IntVar(&o.MaxConcurrentReconciles, "max-concurrent-reconciles", o.MaxConcurrentReconciles, fmt.Sprintf("The number of maximum concurrent reconciles for controller.default is %v", o.MaxConcurrentReconciles))
	cfs.StringVar(&o.Workdir, "workdir", o.Workdir, fmt.Sprintf("The base directory for apiserver.default is %v", o.Workdir))
	cfs.StringVar(&o.UIPath, "ui-path", o.UIPath, fmt.Sprintf("The web ui package path.default is %v", o.UIPath))
	cfs.StringVar(&o.SchemaPath, "schema-path", o.SchemaPath, fmt.Sprintf("The json schema dir path to render web ui.default is %v", o.SchemaPath))
	return fss
}

// Complete for ApiserverOptions
func (o *APIServerOptions) Complete() {
	// do nothing
	if o.MaxConcurrentReconciles == 0 {
		o.MaxConcurrentReconciles = 1
	}
}
