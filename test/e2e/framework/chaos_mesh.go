/*
Copyright 2021 RadonDB.

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

package framework

import (
	"context"
	"fmt"
	"math/rand"

	chaos "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "chaos-mesh.org", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddChaosMeshToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
    scheme.AddKnownTypes(GroupVersion,
        &chaos.PodChaos{},
        &chaos.PodChaosList{},
    )

    metav1.AddToGroupVersion(scheme, GroupVersion)
    return nil
}

func (f Framework) PodFailure(labels map[string]string, duration string, clusterKey types.NamespacedName) error {
	f.WaitRoleAvailable(clusterKey, labels)

	podChaos := f.newPodChaos(labels, chaos.PodFailureAction)
	podChaos.Spec.Duration = &duration
	if err := f.Client.Create(context.TODO(), &podChaos); err != nil {
		return err
	}
	
	f.waitInject(podChaos)
	return nil
}

func (f Framework) ContainerKill(labels map[string]string, containerName string, clusterKey types.NamespacedName) error {
	f.WaitRoleAvailable(clusterKey, labels)

	podChaos := f.newPodChaos(labels, chaos.ContainerKillAction)
	podChaos.Spec.ContainerNames = append(podChaos.Spec.ContainerNames, containerName)
	podChaos.Spec.Selector.Namespaces = append(podChaos.Spec.Selector.Namespaces, clusterKey.Namespace)
	if err := f.Client.Create(context.TODO(), &podChaos); err != nil {
		return err
	}

	f.waitInject(podChaos)
	return nil
}

func (f Framework) PodKill(labels map[string]string, clusterKey types.NamespacedName) error {
	f.WaitRoleAvailable(clusterKey, labels)
	podChaos := f.newPodChaos(labels, chaos.PodKillAction)
	podChaos.Spec.Selector.Namespaces = append(podChaos.Spec.Selector.Namespaces, clusterKey.Namespace)
	if err := f.Client.Create(context.TODO(), &podChaos); err != nil {
		return err
	}
	f.waitInject(podChaos)
	return nil
}

func (f Framework) waitInject(podChao chaos.PodChaos) {
	Eventually(func() bool {
		pc := &chaos.PodChaos{}
		f.Client.Get(context.TODO(), types.NamespacedName{Name: podChao.Name, Namespace: podChao.Namespace}, pc)
		if len(pc.Status.Experiment.Records) == 0 {
			return false
		}
		for _, record := range pc.Status.Experiment.Records {
			if record.Phase == chaos.Injected {
				return true
			}
		}
		return false
	}, FAILOVERTIMEOUT, FAILOVERPOLLING).Should(BeTrue(), "failed to inject in %s", FAILOVERTIMEOUT)
}

func (f Framework) newPodChaos(labels map[string]string, action chaos.PodChaosAction) chaos.PodChaos {
	return chaos.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind: "PodChaos",
			APIVersion: "chaos-mesh.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", string(action), rand.Int31()/1000),
			Namespace: RadondbMysqlE2eNamespace,
		},
		Spec: chaos.PodChaosSpec{
			Action: action,
			ContainerSelector: chaos.ContainerSelector{
				PodSelector: chaos.PodSelector{
					Selector: chaos.PodSelectorSpec{
						GenericSelectorSpec: chaos.GenericSelectorSpec{
							LabelSelectors: labels,
						},
					},
					// Select one to testing.
					Mode: chaos.OneMode,
				},
			},
		},
	}
}

func (f Framework) cleanUpPodChaos() error {
	podChaosList := &chaos.PodChaosList{}
	if err := f.Client.List(context.TODO(), podChaosList, &client.ListOptions{Namespace: RadondbMysqlE2eNamespace}); err != nil {
		return err
	}
	for _, podChaos := range podChaosList.Items {
		if err := f.Client.Delete(context.TODO(), &podChaos); err != nil {
			return err
		}
	}
	return nil
}

func (f Framework) CleanUpChaosMesh() error {
	err := f.cleanUpPodChaos()
	// other chaos
	return err
}