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

package cluster

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var _ = Describe("MySQL Cluster E2E Tests", Label("Cluster"), func() {
	var (
		f          *framework.Framework
		cluster    *apiv1alpha1.MysqlCluster
		clusterKey *types.NamespacedName
		sysbenchOptions  *framework.SysbenchOptions
		leaderLabel map[string]string
		followerLabel map[string]string
		two        = int32(2)
		three      = int32(3)
		five       = int32(5)
	)

	BeforeEach(func() {
		// Singleton
		if f == nil {
			By("Init framework")
			f = &framework.Framework{
				BaseName: "mysqlcluster-e2e",
				Log:      framework.Log,
			}
			f.BeforeEach()
		}
		Expect(f).ShouldNot(BeNil(), "failed to init framework")

		clusterKey = &types.NamespacedName{Namespace: framework.RadondbMysqlE2eNamespace}
		if clusterName := getExistCluster(f); clusterName != "" {
			By("Reuse the exist cluster")
			clusterKey.Name = clusterName
		} else {
			By("Creating a new testing cluster")
			clusterKey.Name = f.InitAClusterForTesting()
		}

		By("Testing the cluster readiness")
		cluster = &apiv1alpha1.MysqlCluster{}
		Expect(f.Client.Get(context.TODO(), *clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
		f.WaitClusterReadiness(cluster)
	})

	// Run the full scale in/out test with label filter: Scale.
	// Run only scale out(2 -> 3 -> 5): Scale out.
	// Run only scale in(5 -> 3 -> 2): Scale in.
	When("Test cluster scale in/out", Label("Scale"), Ordered, func() {
		Context("Scale out", Label("Scale out"), Ordered, func() {
			// 1. Guarantee the initial replicas is 2.
			// 2. Prepare data and run sysbench.
			BeforeAll(func() {
				cluster.Spec.Replicas = &two
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
				f.WaitClusterReadiness(cluster)
				
				sysbenchOptions = &framework.SysbenchOptions{
					Timeout:   10 * time.Minute,
					Threads:   8,
					Tables:    4,
					TableSize: 10000,
				}
				f.PrepareData(cluster, sysbenchOptions)
				f.RunOltpTest(cluster, sysbenchOptions)
			})

			Specify("Replicas: 2 -> 3", func() {
				cluster.Spec.Replicas = &three
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale out finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})

			Specify("Replicas: 3 -> 5", func() {
				cluster.Spec.Replicas = &five
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale out finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})
		})

		Context("Scale in", Label("Scale In"), Ordered, func() {
			// Guarantee the initial replicas is 5.
			BeforeAll(func() {
				cluster.Spec.Replicas = &five
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
				f.WaitClusterReadiness(cluster)
			})

			Specify("Replicas: 5 -> 3", func() {
				cluster.Spec.Replicas = &three
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale in finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})
			Specify("Replicas: 3 -> 2", func() {
				cluster.Spec.Replicas = &two
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale in finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})
		})
	})

	When("Testing cluster HA", Label("HA"), func() {
		leaderLabel = map[string]string{
			"role": string(utils.Leader),
		}
		followerLabel = map[string]string{
			"role": string(utils.Follower),
		}

		Context("Kill container", Label("Kill container", "Two"), func ()  {
			Specify("Kill Leader`s MySQL", Label("MySQL", "Leader"), func() {
				Expect(f.ContainerKill(leaderLabel, "mysql", *clusterKey)).To(Succeed())
				fmt.Println("failover time: ", waitClusterFailover(f, *clusterKey))
			})

			Specify("Kill Leader`s Xenon", Label("Xenon", "Leader"), func() {
				Expect(f.ContainerKill(leaderLabel, "xenon", *clusterKey)).To(Succeed())
				fmt.Println("failover time: ", waitClusterFailover(f, *clusterKey))
			})

			Specify("Kill Follower`s MySQL", Label("MySQL", "Follower"), func() {
				Expect(f.ContainerKill(followerLabel, "mysql", *clusterKey)).To(Succeed())
				fmt.Println("failover time: ", waitClusterFailover(f, *clusterKey))
			})

			Specify("Kill Follower`s Xenon", Label("Xenon", "Follower"), func() {
				Expect(f.ContainerKill(followerLabel, "xenon", *clusterKey)).To(Succeed())
				fmt.Println("failover time: ", waitClusterFailover(f, *clusterKey))
			})
		})

		Context("Kill pod", Label("Kill pod", "Two"), func ()  {
			Specify("Kill Leader", Label("Leader"), func() {
				Expect(f.PodKill(leaderLabel, *clusterKey)).To(Succeed())
				fmt.Println("failover time: ", waitClusterFailover(f, *clusterKey))
			})

			Specify("Kill Follower", Label("Follower"), func() {
				Expect(f.PodKill(leaderLabel, *clusterKey)).To(Succeed())
				fmt.Println("failover time: ", waitClusterFailover(f, *clusterKey))
			})
		})
	})
})

func getExistCluster(f *framework.Framework) string {
	existClusters := &apiv1alpha1.MysqlClusterList{}
	Expect(f.Client.List(context.TODO(), existClusters, &client.ListOptions{
		Namespace: framework.RadondbMysqlE2eNamespace,
	})).To(Succeed(), "failed to list clusters")

	if len(existClusters.Items) > 0 {
		return existClusters.Items[0].Name
	}
	return ""
}

func waitClusterFailover(f *framework.Framework, clusterKey types.NamespacedName) time.Duration {
	// Waiting for cluster discovery error.
	Eventually(func() bool {
		return isClusterReadiness(f, clusterKey)
	}, framework.FAILOVERTIMEOUT, framework.FAILOVERPOLLING).ShouldNot(BeTrue(), "cluster cant discover error")

	startTime := time.Now()
	// Waiting for cluster recovery.
	Eventually(func() bool {
		return isClusterReadiness(f, clusterKey)
	}, framework.FAILOVERTIMEOUT, framework.FAILOVERPOLLING).Should(BeTrue(), "cluster cannot failover in %s", framework.FAILOVERTIMEOUT)
	return time.Since(startTime)
}

func isClusterReadiness(f *framework.Framework, clusterKey types.NamespacedName) bool {
	cluster := &apiv1alpha1.MysqlCluster{}
	f.Client.Get(context.TODO(), clusterKey, cluster)
	if cluster.Status.State == apiv1alpha1.ClusterReadyState && framework.IsXenonReadiness(cluster) {
		return true
	}
	return false
}
