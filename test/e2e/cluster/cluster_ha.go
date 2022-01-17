// /*
// Copyright 2021 RadonDB.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

// package cluster

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// 	"k8s.io/apimachinery/pkg/types"

// 	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
// 	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
// )

// var _ = Describe("MySQL Cluster E2E Tests", func() {
// 	leaderLabel := map[string]string{
// 		"role": "leader",
// 	}

// 	followerLabel := map[string]string{
// 		"role": "follower",
// 	}

// 	var (
// 		f          *framework.Framework
// 		name       string
// 		cluster    *apiv1alpha1.MysqlCluster
// 		clusterKey *types.NamespacedName
// 		three      = int32(3)
// 	)

// 	BeforeEach(func() {
// 		// Singleton
// 		if f == nil {
// 			By("Init framework")
// 			f = &framework.Framework{
// 				BaseName: "mysqlcluster-e2e",
// 				Log:      framework.Log,
// 			}
// 			f.BeforeEach()
// 		}
// 		Expect(f).ShouldNot(BeNil(), "failed to init framework")

// 		clusterKey = &types.NamespacedName{Namespace: framework.RadondbMysqlE2eNamespace}
// 		if clusterName := getExistCluster(f); clusterName != "" {
// 			By("Reuse the exist cluster")
// 			clusterKey.Name = clusterName
// 		} else {
// 			By("Creating a new testing cluster")
// 			clusterKey.Name = f.InitAClusterForTesting()
// 		}

// 		By("Testing the cluster readiness")
// 		cluster = &apiv1alpha1.MysqlCluster{}
// 		Expect(f.Client.Get(context.TODO(), *clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
// 		f.WaitClusterReadiness(cluster)
// 	})

// 	AfterEach(func() {
// 		By("Do after each")
// 	})

// 	PIt("HA test: kill container", func() {
// 		By("Replicas 2: kill leader`s xenon")
// 		Expect(f.ContainerKill(leaderLabel, "xenon", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 2: kill follower`s xenon")
// 		Expect(f.ContainerKill(followerLabel, "xenon", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))
		
// 		By("Replicas 2: kill leader`s mysql")
// 		Expect(f.ContainerKill(leaderLabel, "mysql", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 2: kill follower`s mysql")
// 		Expect(f.ContainerKill(followerLabel, "mysql", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("scale out: 2 -> 3")
// 		Expect(f.Client.Get(context.TODO(), clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
// 		cluster.Spec.Replicas = &three
// 		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
// 		fmt.Println("scale time: ", waitClusterReadiness(f, cluster))

// 		By("Replicas 3: kill leader`s xenon")
// 		Expect(f.ContainerKill(leaderLabel, "xenon", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 3: kill one follower`s xenon")
// 		Expect(f.ContainerKill(followerLabel, "xenon", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))
		
// 		By("Replicas 3: kill leader`s mysql")
// 		Expect(f.ContainerKill(leaderLabel, "mysql", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 3: kill one follower`s mysql")
// 		Expect(f.ContainerKill(followerLabel, "mysql", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))
// 	})

// 	It("HA test: kill pod", func() {
// 		By("Replicas 2: kill leader pod")
// 		Expect(f.PodKill(leaderLabel, clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 2: kill follower pod")
// 		Expect(f.PodKill(followerLabel, clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("scale out: 2 -> 3")
// 		Expect(f.Client.Get(context.TODO(), clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
// 		cluster.Spec.Replicas = &three
// 		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
// 		fmt.Println("scale time: ", waitClusterReadiness(f, cluster))

// 		By("Replicas 3: kill one follower pod")
// 		Expect(f.PodKill(followerLabel, clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 3: kill leader pod")
// 		Expect(f.PodKill(leaderLabel, clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))
// 	})

// 	PIt("HA test: pod failure", func() {
// 		By("Replicas 2: leader pod failure 10s")
// 		Expect(f.PodFailure(leaderLabel, "10s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 2: follower pod failure 10s")
// 		Expect(f.PodFailure(followerLabel, "10s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 2: leader pod failure 30s")
// 		Expect(f.PodFailure(leaderLabel, "30s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 2: follower pod failure 30s")
// 		Expect(f.PodFailure(followerLabel, "30s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("scale out: 2 -> 3")
// 		Expect(f.Client.Get(context.TODO(), clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
// 		cluster.Spec.Replicas = &three
// 		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
// 		fmt.Println("scale time: ", waitClusterReadiness(f, cluster))

// 		By("Replicas 3: leader pod failure 10s")
// 		Expect(f.PodFailure(leaderLabel, "10s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 3: one of follower pod failure 10s")
// 		Expect(f.PodFailure(followerLabel, "10s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 3: leader pod failure 30s")
// 		Expect(f.PodFailure(leaderLabel, "30s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))

// 		By("Replicas 3: one of follower pod failure 30s")
// 		Expect(f.PodFailure(followerLabel, "30s", clusterKey)).To(Succeed())
// 		fmt.Println("failover time: ", waitClusterFailover(f, clusterKey))
// 	})

// })


// func waitClusterFailover(f *framework.Framework, clusterKey types.NamespacedName) time.Duration {
// 	// Waiting for cluster discovery error.
// 	Eventually(func() bool {
// 		return isClusterReadiness(f, clusterKey)
// 	}, framework.FAILOVERTIMEOUT, framework.FAILOVERPOLLING).ShouldNot(BeTrue(), "cluster cant discover error")

// 	startTime := time.Now()
// 	// Waiting for cluster recovery.
// 	Eventually(func() bool {
// 		return isClusterReadiness(f, clusterKey)
// 	}, framework.FAILOVERTIMEOUT, framework.FAILOVERPOLLING).Should(BeTrue(), "cluster cannot failover in %s", framework.FAILOVERTIMEOUT)
// 	return time.Since(startTime)
// }

// func isClusterReadiness(f *framework.Framework, clusterKey types.NamespacedName) bool {
// 	cluster := &apiv1alpha1.MysqlCluster{}
// 	f.Client.Get(context.TODO(), clusterKey, cluster)
// 	if cluster.Status.State == apiv1alpha1.ClusterReadyState && framework.IsXenonReadiness(cluster) {
// 		return true
// 	}
// 	return false
// }

