/*


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

package controllers

import (
	"context"

	"github.com/alibaba/sentinel-golang/core/system"
	datasourcev1 "github.com/alibaba/sentinel-golang/ext/datasource/k8s/api/v1"
	"github.com/alibaba/sentinel-golang/logging"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SystemRulesReconciler reconciles a SystemRules object
type SystemRulesReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	EffectiveCrName string
}

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=systemrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=systemrules/status,verbs=get;update;patch

func (r *SystemRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logging.Info("receive SystemRules", "namespace", req.NamespacedName.String())

	if req.Name != r.EffectiveCrName {
		logging.Warn("ignore unregister cr.", "ns", req.Namespace, "crName", req.Name)
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	systemRulesCR := &datasourcev1.SystemRules{}
	if err := r.Get(ctx, req.NamespacedName, systemRulesCR); err != nil {
		logging.Error(err, "Fail to get datasourcev1.SystemRules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}

	logging.Info("Receive datasourcev1.SystemRules", "rules:", systemRulesCR.Spec.Rules)

	systemRules := r.assembleSystemRules(systemRulesCR)
	_, err := system.LoadRules(systemRules)
	if err != nil {
		logging.Error(err, "Fail to Load system.Rules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	return ctrl.Result{}, nil
}

func (r *SystemRulesReconciler) assembleSystemRules(rs *datasourcev1.SystemRules) []*system.Rule {
	ret := make([]*system.Rule, 0, len(rs.Spec.Rules))
	for _, rule := range rs.Spec.Rules {
		systemRule := &system.Rule{
			MetricType:   0,
			TriggerCount: 0,
			Strategy:     0,
		}

		switch rule.MetricType {
		case "Load":
			systemRule.MetricType = system.Load
			systemRule.TriggerCount = float64(rule.TriggerCount) / 100
		case "AvgRT":
			systemRule.MetricType = system.AvgRT
			systemRule.TriggerCount = float64(rule.TriggerCount)
		case "Concurrency":
			systemRule.MetricType = system.Concurrency
			systemRule.TriggerCount = float64(rule.TriggerCount)
		case "InboundQPS":
			systemRule.MetricType = system.InboundQPS
			systemRule.TriggerCount = float64(rule.TriggerCount)
		case "CpuUsage":
			systemRule.MetricType = system.CpuUsage
			systemRule.TriggerCount = float64(rule.TriggerCount) / 100
		default:
			logging.Error("unsupported MetricType for system.Rule", "metricType", rule.MetricType)
			continue
		}
		switch rule.Strategy {
		case "NoAdaptive":
			systemRule.Strategy = system.NoAdaptive
		case "BBR":
			systemRule.Strategy = system.BBR
		default:
			logging.Error("unsupported Strategy for system.Rule", "strategy", rule.Strategy)
			continue
		}
		ret = append(ret, systemRule)
	}
	return ret
}

func (r *SystemRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.SystemRules{}).
		Complete(r)
}
