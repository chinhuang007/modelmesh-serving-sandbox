// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// IBM Confidential
// OCO Source Materials
// 5900-AEO
//
// Copyright IBM Corp. 2021
//
// The source code for this program is not published or otherwise
// divested of its trade secrets, irrespective of what has been
// deposited with the U.S. Copyright Office.
// ------------------------------------------------------ {COPYRIGHT-END} ---
package mmesh

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type mockClient struct {
	t       *testing.T
	getfunc func(context.Context, client.ObjectKey, *v1.Endpoints) error
}

func (m mockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	assert.NotNil(m.t, ctx)
	assert.IsType(m.t, &v1.Endpoints{}, obj)
	assert.Equal(m.t, "modelmesh-serving", key.Name)
	assert.Equal(m.t, "namespace", key.Namespace)
	return m.getfunc(ctx, key, obj.(*v1.Endpoints))
}

type mockCC struct {
	t          *testing.T
	updatefunc func(state resolver.State)
}

func (m mockCC) UpdateState(state resolver.State) {
	assert.NotNil(m.t, state)
	fmt.Printf("updatestate called: %v\n", state)
	m.updatefunc(state)
}

// Test for basic functionality
func Test_KubeResolver_AddRemove(t *testing.T) {
	mClient := mockClient{t: t}
	mClient.getfunc = func(ctx context.Context, key client.ObjectKey, ep *v1.Endpoints) error {
		ep.Subsets = []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{IP: "1.2.3.4"},
				},
				NotReadyAddresses: []v1.EndpointAddress{},
				Ports: []v1.EndpointPort{
					{Name: "grpc", Port: 8033},
					{Name: "prometheus", Port: 2112},
				},
			},
		}
		return nil
	}

	kr := makeKubeResolver("namespace", mClient)

	mCC := mockCC{}
	updateStateCalled := false
	mCC.updatefunc = func(state resolver.State) {
		updateStateCalled = true
		assert.Len(t, state.Addresses, 1)
		assert.Equal(t, "1.2.3.4:8033", state.Addresses[0].Addr)
	}

	fmt.Println("Build r1")
	r1, err := kr.Build(resolver.Target{Scheme: "kube", Endpoint: "modelmesh-serving:8033"}, mCC, resolver.BuildOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, r1)
	assert.True(t, updateStateCalled)
	updateStateCalled = false

	reconcile(t, kr)
	assert.True(t, updateStateCalled)
	updateStateCalled = false

	mCC2 := mockCC{}
	updateState2Called := false
	mCC2.updatefunc = func(state resolver.State) {
		updateState2Called = true
		assert.Len(t, state.Addresses, 1)
		assert.Equal(t, "1.2.3.4:8033", state.Addresses[0].Addr)
	}

	fmt.Println("Build r2")
	r2, err := kr.Build(resolver.Target{Scheme: "kube", Endpoint: "modelmesh-serving:8033"}, mCC2, resolver.BuildOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, r2)
	assert.False(t, updateStateCalled)
	assert.True(t, updateState2Called)
	updateState2Called = false

	reconcile(t, kr)
	assert.True(t, updateStateCalled)
	assert.True(t, updateState2Called)
	updateStateCalled, updateState2Called = false, false

	fmt.Println("Close r1")
	r1.Close()

	reconcile(t, kr)
	assert.False(t, updateStateCalled)
	assert.True(t, updateState2Called)
	updateStateCalled, updateState2Called = false, false

	fmt.Println("Close r2")
	r2.Close()

	reconcile(t, kr)
	assert.False(t, updateStateCalled)
	assert.False(t, updateState2Called)
	updateStateCalled, updateState2Called = false, false
}

func reconcile(t *testing.T, kr *KubeResolver) {
	fmt.Println("Reconcile")
	_, err := kr.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{
		Namespace: "namespace", Name: "modelmesh-serving",
	}})
	assert.Nil(t, err)
}

func makeKubeResolver(namespace string, client client.Client) *KubeResolver {
	return &KubeResolver{
		namespace: namespace, Client: client,
		resolvers: make(map[string][]*serviceResolver, 2),
		logger:    zap.New(zap.UseDevMode(true)).WithName("KubeResolver"),
	}
}

// Unused mock funcs

func (m mockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) Status() client.StatusWriter {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) Scheme() *runtime.Scheme {
	m.t.Error("should not be called")
	return nil
}

func (m mockClient) RESTMapper() meta.RESTMapper {
	m.t.Error("should not be called")
	return nil
}

func (m mockCC) ReportError(err error) {
	m.t.Error("should not be called")
}

func (m mockCC) NewAddress(addresses []resolver.Address) {
	m.t.Error("should not be called")
}

func (m mockCC) NewServiceConfig(serviceConfig string) {
	m.t.Error("should not be called")
}

func (m mockCC) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	m.t.Error("should not be called")
	return nil
}
