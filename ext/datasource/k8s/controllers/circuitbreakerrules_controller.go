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
	"errors"
	datasourcev1 "sentinel-go-k8s-crd-datasource/api/v1"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CircuitBreakerRulesReconciler reconciles a CircuitBreakerRules object
type CircuitBreakerRulesReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	SlowRequestRatioStrategy string = "SlowRequestRatio"
	ErrorRatioStrategy              = "ErrorRatio"
	ErrorCountStrategy              = "ErrorCount"
)

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=circuitbreakerrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=circuitbreakerrules/status,verbs=get;update;patch

func (r *CircuitBreakerRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("circuit breaker rules", req.NamespacedName)

	cbRulesCR := &datasourcev1.CircuitBreakerRules{}
	if err := r.Get(ctx, req.NamespacedName, cbRulesCR); err != nil {
		log.Error(err, "Fail to get datasourcev1.CircuitBreakerRules.")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	log.Info("Receive datasourcev1.CircuitBreakerRules", "rules:", cbRulesCR.Spec.Rules)

	cbRules := r.assembleCircuitBreakerRules(cbRulesCR)
	_, err := circuitbreaker.LoadRules(cbRules)
	if err != nil {
		log.Error(err, "Fail to Load circuitbreaker.Rules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	return ctrl.Result{}, nil
}

func (r *CircuitBreakerRulesReconciler) assembleCircuitBreakerRules(rs *datasourcev1.CircuitBreakerRules) []*circuitbreaker.Rule {
	ret := make([]*circuitbreaker.Rule, 0, len(rs.Spec.Rules))

	for _, rule := range rs.Spec.Rules {
		cbRule := &circuitbreaker.Rule{
			Id:               rule.Id,
			Resource:         rule.Resource,
			RetryTimeoutMs:   uint32(rule.RetryTimeoutMs),
			MinRequestAmount: uint64(rule.MinRequestAmount),
			StatIntervalMs:   uint32(rule.StatIntervalMs),
			MaxAllowedRtMs:   uint64(rule.MaxAllowedRtMs),
		}
		switch rule.Strategy {
		case SlowRequestRatioStrategy:
			cbRule.Strategy = circuitbreaker.SlowRequestRatio
			cbRule.Threshold = float64(rule.Threshold) / 100
		case ErrorRatioStrategy:
			cbRule.Strategy = circuitbreaker.ErrorRatio
			cbRule.Threshold = float64(rule.Threshold) / 100
		case ErrorCountStrategy:
			cbRule.Strategy = circuitbreaker.ErrorCount
			cbRule.Threshold = float64(rule.Threshold)
		default:
			r.Log.Error(errors.New("unsupported circuit breaker strategy"), rule.Strategy)
			continue
		}

		ret = append(ret, cbRule)
	}
	return ret
}

func (r *CircuitBreakerRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.CircuitBreakerRules{}).
		Complete(r)
}
