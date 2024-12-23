//go:build builtin
// +build builtin

/*
Copyright 2024 The KubeSphere Authors.

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
	"fmt"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kubesphere/kubekey/v4/builtin/core"
	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

func complateInventory() error {
	data, err := core.Defaults.ReadFile("inventory/localhost.yaml")
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, options.Inventory); err != nil {
		return err
	}

	return nil
}

func complateConfig(kubeVersion string) error {
	data, err := core.Defaults.ReadFile(fmt.Sprintf("config/%s.yaml", kubeVersion))
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, options.Config); err != nil {
		return err
	}

	return nil
}
