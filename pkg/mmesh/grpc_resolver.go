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

// not needed because we resolve whenever there are any endpoint changes
func (r *serviceResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *serviceResolver) Close() {
	r.owner.resolvers.Delete(r.name)
}

// Should only be called once
func InitGrpcResolver(namespace string, mgr ctrl.Manager) (*KubeResolver, error) {
	kr := &KubeResolver{namespace: namespace, Client: mgr.GetClient()}
	err := ctrl.NewControllerManagedBy(mgr).For(&corev1.Endpoints{}).Complete(kr)
	if err != nil {
		return nil, err
	}
	resolver.Register(kr)
	return kr, nil
}

// KubeResolver is a ResolverBuilder and a Reconciler
type KubeResolver struct {
	client.Client
	namespace string

	// Map of resolvers in use
	resolvers sync.Map
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

	r := serviceResolver{name: parts[0], port: parts[1], cc: cc, owner: kr}
	kr.resolvers.Store(r.name, r)
	_, err := kr.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: kr.namespace, Name: r.name},
	})
	return &r, err
}

func (*KubeResolver) Scheme() string {
	return KUBE_SCHEME
}

func (kr *KubeResolver) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace != kr.namespace {
		return ctrl.Result{}, nil
	}
	v, ok := kr.resolvers.Load(req.Name)
	if !ok {
		return ctrl.Result{}, nil
	}
	r := v.(serviceResolver)
	endpoints := &corev1.Endpoints{}
	err := kr.Get(ctx, req.NamespacedName, endpoints)
	if errors.IsNotFound(err) {
		r.cc.ReportError(fmt.Errorf("kube Service %s not found", req.Name))
		return ctrl.Result{}, nil // not an error from reconciler pov
	}
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error obtaining endpoints for service %s: %w", req.Name, err)
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
