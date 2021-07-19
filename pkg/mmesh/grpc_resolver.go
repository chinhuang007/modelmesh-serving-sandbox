// Copyright 2021 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mmesh

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/go-logr/logr"

	"google.golang.org/grpc/resolver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const KUBE_SCHEME = "kube"

type serviceResolver struct {
	name  string
	port  string
	owner *KubeResolver
	cc    resolver.ClientConn
}

// ResolveNow not needed because we resolve whenever there are any endpoint changes
func (sr *serviceResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (sr *serviceResolver) Close() {
	kr := sr.owner
	kr.lock.Lock()
	defer kr.lock.Unlock()
	list, ok := kr.resolvers[sr.name]
	if ok {
		for i, r := range list {
			if r == sr {
				if last := len(list) - 1; last <= 0 {
					delete(kr.resolvers, sr.name) // remove entry if last one
				} else {
					list[i] = list[last]
					kr.resolvers[sr.name] = list[:last]
				}
				kr.logger.V(1).Info("Removed resolver", "name", sr.name)
				return
			}
		}
	}
	kr.logger.V(1).Info("Close called on unrecognized resolver", "name", sr.name)
}

// InitGrpcResolver should only be called once
func InitGrpcResolver(namespace string, mgr ctrl.Manager) (*KubeResolver, error) {
	kr := &KubeResolver{
		namespace: namespace, Client: mgr.GetClient(),
		resolvers: make(map[string][]*serviceResolver, 2),
		logger:    ctrl.Log.WithName("KubeResolver"),
	}
	err := ctrl.NewControllerManagedBy(mgr).For(&corev1.Endpoints{}).Complete(kr)
	if err != nil {
		return nil, err
	}
	resolver.Register(kr)
	kr.logger.Info("Registered KubeResolver with kubebuilder and gRPC")
	return kr, nil
}

// KubeResolver is a ResolverBuilder and a Reconciler
type KubeResolver struct {
	client.Client
	namespace string

	// Map of resolvers in use
	resolvers map[string][]*serviceResolver
	lock      sync.Mutex

	logger logr.Logger
}

func (kr *KubeResolver) Build(target resolver.Target, cc resolver.ClientConn,
	_ resolver.BuildOptions) (resolver.Resolver, error) {
	if target.Scheme != KUBE_SCHEME {
		return nil, fmt.Errorf("unsupported scheme: %s", target.Scheme)
	}
	parts := strings.Split(target.Endpoint, ":")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return nil, fmt.Errorf("target must be of form: %s:///servicename:port", target.Scheme)
	}

	r := &serviceResolver{name: parts[0], port: parts[1], cc: cc, owner: kr}
	kr.lock.Lock()
	defer kr.lock.Unlock()
	list, ok := kr.resolvers[r.name]
	singleton := []*serviceResolver{r}
	if ok {
		kr.resolvers[r.name] = append(list, r)
	} else {
		kr.resolvers[r.name] = singleton
	}
	log := r.owner.logger.V(1)
	log.Info("Built new resolver", "target", target, "name", r.name)
	// Initialize resolver state before returning via a synchronous reconciliation
	_, err := kr.reconcile(context.TODO(),
		ctrl.Request{NamespacedName: types.NamespacedName{Namespace: kr.namespace, Name: r.name}},
		singleton, log)
	return r, err
}

func (*KubeResolver) Scheme() string {
	return KUBE_SCHEME
}

func (kr *KubeResolver) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace != kr.namespace {
		return ctrl.Result{}, nil
	}
	log := kr.logger.WithName("Reconcile").V(1)
	kr.lock.Lock()
	defer kr.lock.Unlock()
	if list, ok := kr.resolvers[req.Name]; ok {
		return kr.reconcile(ctx, req, list, log)
	}
	log.Info("Ignoring event for Endpoints with no resolver", "endpoints", req.Name)
	return ctrl.Result{}, nil
}

// called under lock
func (kr *KubeResolver) reconcile(ctx context.Context, req ctrl.Request,
	list []*serviceResolver, log logr.Logger) (ctrl.Result, error) {
	endpoints := &corev1.Endpoints{}
	err := kr.Get(ctx, req.NamespacedName, endpoints)
	if err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error obtaining endpoints for service %s: %w", req.Name, err)
		} else {
			log.Info("Endpoints not found", "endpoints", req.Name)
		}
	}
	for _, r := range list {
		if errors.IsNotFound(err) {
			r.cc.ReportError(fmt.Errorf("kube Service %s not found", req.Name))
			continue // not an error from reconciler pov
		}
		var addrs []resolver.Address
		for _, s := range endpoints.Subsets {
			if p := hasTargetPort(&s, r.port); p > 0 {
				for _, ea := range s.Addresses {
					addrs = append(addrs, resolver.Address{Addr: fmt.Sprintf("%s:%d", ea.IP, p)})
				}
			}
		}
		r.cc.UpdateState(resolver.State{Addresses: addrs})
		log.Info("Updated resolver state with new endpoints", "endpoints", req.Name, "count", len(addrs))
	}
	return ctrl.Result{}, nil
}

// returns int32 port number if port string matches name or number of port in EndpointSubset
func hasTargetPort(s *corev1.EndpointSubset, port string) int32 {
	for _, p := range s.Ports {
		if p.Name == port || strconv.Itoa(int(p.Port)) == port {
			return p.Port
		}
	}
	return -1
}
