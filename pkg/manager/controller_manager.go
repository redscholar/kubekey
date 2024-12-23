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

package manager

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/controllers"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

type controllerManager struct {
	MaxConcurrentReconciles int
	LeaderElection          bool
}

// Run controllerManager, run controller in kubernetes
func (m controllerManager) Run(ctx context.Context) error {
	ctrl.SetLogger(klog.NewKlogr())
	restconfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Infof("kubeconfig in empty, store resources local")
		restconfig = &rest.Config{}
	}
	restconfig, err = proxy.NewConfig(restconfig)
	if err != nil {
		return fmt.Errorf("could not get rest config: %w", err)
	}

	mgr, err := ctrl.NewManager(restconfig, ctrl.Options{
		Scheme:                 _const.Scheme,
		LeaderElection:         m.LeaderElection,
		LeaderElectionID:       "controller-leader-election-kk",
		HealthProbeBindAddress: ":9440",
	})
	if err != nil {
		return fmt.Errorf("could not create controller manager: %w", err)
	}

	if err := m.register(mgr); err != nil {
		return err
	}

	return mgr.Start(ctx)
}

func (m controllerManager) register(mgr ctrl.Manager) error {
	o := &ctrlcontroller.Options{
		MaxConcurrentReconciles: m.MaxConcurrentReconciles,
	}
	for _, c := range controllers.Controllers {
		if err := c.SetupWithManager(mgr, *o); err != nil {
			klog.ErrorS(err, "register controller error", "controller", c.Name())

			return err
		}
	}

	return nil
}
