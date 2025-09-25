/*
Copyright 2025.

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

package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Helper function to run oc commands
func runOC(args ...string) (string, error) {
	cmd := exec.Command("oc", args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// Helper function to run oc commands and expect success
func runOCExpectSuccess(args ...string) string {
	output, err := runOC(args...)
	Expect(err).NotTo(HaveOccurred(), "oc command failed: %s\nOutput: %s", strings.Join(args, " "), output)
	return output
}

var _ = Describe("Kueue Operator Installation and Health Tests", func() {
	const (
		kueueOperatorNamespace  = "openshift-kueue-operator"
		kueueOperatorDeployment = "openshift-kueue-operator"
		kueueCRDName            = "kueues.kueue.openshift.io"
		timeout                 = 5 * time.Minute
		interval                = 10 * time.Second
	)

	BeforeEach(func() {
		// Verify oc command is available
		_, err := exec.LookPath("oc")
		Expect(err).NotTo(HaveOccurred(), "oc command not found in PATH")

		// Verify we're logged in to OpenShift
		_, err = runOC("whoami")
		Expect(err).NotTo(HaveOccurred(), "Not logged in to OpenShift cluster")
	})

	Describe("Operator Installation Verification", func() {
		It("should have the kueue operator namespace", func() {
			By("Verifying the kueue operator namespace exists")

			Eventually(func() error {
				_, err := runOC("get", "namespace", kueueOperatorNamespace)
				if err != nil {
					return fmt.Errorf("namespace %s not found: %v", kueueOperatorNamespace, err)
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})

		It("should have the kueue CRD installed", func() {
			By("Checking that the Kueue CRD is present and established")

			Eventually(func() error {
				output, err := runOC("get", "crd", kueueCRDName, "-o", "jsonpath={.status.conditions[?(@.type=='Established')].status}")
				if err != nil {
					return fmt.Errorf("CRD %s not found: %v", kueueCRDName, err)
				}

				if strings.TrimSpace(output) != "True" {
					return fmt.Errorf("CRD %s is not established, status: %s", kueueCRDName, output)
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})

		It("should have a healthy operator deployment", func() {
			By("Verifying the kueue operator deployment is ready")

			Eventually(func() error {
				// Check if deployment exists and get status
				output, err := runOC("get", "deployment", kueueOperatorDeployment, "-n", kueueOperatorNamespace, "-o", "json")
				if err != nil {
					return fmt.Errorf("deployment %s not found: %v", kueueOperatorDeployment, err)
				}

				var deployment map[string]interface{}
				if err := json.Unmarshal([]byte(output), &deployment); err != nil {
					return fmt.Errorf("failed to parse deployment JSON: %v", err)
				}

				status := deployment["status"].(map[string]interface{})
				replicas := int(status["replicas"].(float64))
				readyReplicas := int(status["readyReplicas"].(float64))

				if replicas == 0 {
					return fmt.Errorf("deployment has 0 replicas configured")
				}

				if readyReplicas != replicas {
					return fmt.Errorf("deployment not ready: %d/%d replicas ready", readyReplicas, replicas)
				}

				// Check if deployment is available
				available, err := runOC("get", "deployment", kueueOperatorDeployment, "-n", kueueOperatorNamespace, "-o", "jsonpath={.status.conditions[?(@.type=='Available')].status}")
				if err != nil {
					return fmt.Errorf("failed to check deployment availability: %v", err)
				}

				if strings.TrimSpace(available) != "True" {
					return fmt.Errorf("deployment not available")
				}

				return nil
			}, timeout, interval).Should(Succeed())
		})

		It("should have running and ready operator pods", func() {
			By("Checking that operator pods are running and ready")

			Eventually(func() error {
				// Get pods with the operator label
				output, err := runOC("get", "pods", "-n", kueueOperatorNamespace, "-l", "name="+kueueOperatorDeployment, "-o", "json")
				if err != nil {
					return fmt.Errorf("failed to list pods: %v", err)
				}

				var podList map[string]interface{}
				if err := json.Unmarshal([]byte(output), &podList); err != nil {
					return fmt.Errorf("failed to parse pods JSON: %v", err)
				}

				items := podList["items"].([]interface{})
				if len(items) == 0 {
					return fmt.Errorf("no operator pods found")
				}

				readyPods := 0
				for _, item := range items {
					pod := item.(map[string]interface{})
					status := pod["status"].(map[string]interface{})

					// Skip pods that are being deleted
					if metadata := pod["metadata"].(map[string]interface{}); metadata["deletionTimestamp"] != nil {
						continue
					}

					// Check pod phase
					if status["phase"].(string) != "Running" {
						continue
					}

					// Check container readiness
					if containerStatuses, ok := status["containerStatuses"].([]interface{}); ok {
						allReady := true
						for _, cs := range containerStatuses {
							containerStatus := cs.(map[string]interface{})
							if !containerStatus["ready"].(bool) {
								allReady = false
								break
							}
						}
						if allReady {
							readyPods++
						}
					}
				}

				if readyPods == 0 {
					return fmt.Errorf("no ready operator pods found")
				}

				return nil
			}, timeout, interval).Should(Succeed())
		})

		It("should have proper RBAC configuration", func() {
			By("Verifying service account exists")
			Eventually(func() error {
				_, err := runOC("get", "serviceaccount", kueueOperatorDeployment, "-n", kueueOperatorNamespace)
				if err != nil {
					return fmt.Errorf("service account %s not found: %v", kueueOperatorDeployment, err)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("Verifying cluster role exists")
			Eventually(func() error {
				_, err := runOC("get", "clusterrole", kueueOperatorDeployment)
				if err != nil {
					return fmt.Errorf("cluster role %s not found: %v", kueueOperatorDeployment, err)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("Verifying cluster role binding exists")
			Eventually(func() error {
				_, err := runOC("get", "clusterrolebinding", kueueOperatorDeployment)
				if err != nil {
					return fmt.Errorf("cluster role binding %s not found: %v", kueueOperatorDeployment, err)
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})
	})

	Describe("Operator Health Verification", func() {
		It("should have healthy operator pod logs", func() {
			By("Checking operator logs for critical errors")

			Eventually(func() error {
				// Get running pods
				podNames, err := runOC("get", "pods", "-n", kueueOperatorNamespace, "-l", "name="+kueueOperatorDeployment, "--field-selector=status.phase=Running", "-o", "jsonpath={.items[*].metadata.name}")
				if err != nil {
					return fmt.Errorf("failed to get running pods: %v", err)
				}

				if strings.TrimSpace(podNames) == "" {
					return fmt.Errorf("no running operator pods found")
				}

				pods := strings.Fields(podNames)
				for _, podName := range pods {
					logs, err := runOC("logs", podName, "-n", kueueOperatorNamespace, "-c", kueueOperatorDeployment, "--tail=100")
					if err != nil {
						continue // Try next pod
					}

					// Check for critical errors
					errorPatterns := []string{
						"panic:",
						"fatal error:",
						"failed to start manager",
						"unable to create controller",
					}

					for _, pattern := range errorPatterns {
						if strings.Contains(strings.ToLower(logs), strings.ToLower(pattern)) {
							return fmt.Errorf("found critical error in logs: %s", pattern)
						}
					}

					// Look for positive indicators that operator is working
					if strings.Contains(logs, "Starting manager") ||
						strings.Contains(logs, "Starting workers") ||
						strings.Contains(logs, "controller-runtime") {
						return nil
					}
				}

				return fmt.Errorf("no healthy operator logs found")
			}, timeout, interval).Should(Succeed())
		})

		It("should expose metrics endpoint", func() {
			By("Verifying metrics port is configured on operator pods")

			Eventually(func() error {
				// Check if metrics port is exposed in the deployment
				output, err := runOC("get", "deployment", kueueOperatorDeployment, "-n", kueueOperatorNamespace, "-o", "jsonpath={.spec.template.spec.containers[?(@.name=='"+kueueOperatorDeployment+"')].ports[?(@.name=='metrics')].containerPort}")
				if err != nil {
					return fmt.Errorf("failed to check metrics port: %v", err)
				}

				if strings.TrimSpace(output) != "60000" {
					return fmt.Errorf("metrics port not found or incorrect, expected 60000, got: %s", output)
				}

				return nil
			}, timeout, interval).Should(Succeed())
		})

		It("should have correct security configuration", func() {
			By("Verifying operator runs with security constraints")

			Eventually(func() error {
				// Check pod security context
				runAsNonRoot, err := runOC("get", "deployment", kueueOperatorDeployment, "-n", kueueOperatorNamespace, "-o", "jsonpath={.spec.template.spec.securityContext.runAsNonRoot}")
				if err != nil {
					return fmt.Errorf("failed to check pod security context: %v", err)
				}

				if strings.TrimSpace(runAsNonRoot) != "true" {
					return fmt.Errorf("pod not configured to run as non-root")
				}

				// Check container security context - privilege escalation
				allowPrivilegeEscalation, err := runOC("get", "deployment", kueueOperatorDeployment, "-n", kueueOperatorNamespace, "-o", "jsonpath={.spec.template.spec.containers[?(@.name=='"+kueueOperatorDeployment+"')].securityContext.allowPrivilegeEscalation}")
				if err != nil {
					return fmt.Errorf("failed to check privilege escalation: %v", err)
				}

				if strings.TrimSpace(allowPrivilegeEscalation) != "false" {
					return fmt.Errorf("privilege escalation not disabled")
				}

				// Check container security context - read-only root filesystem
				readOnlyRootFilesystem, err := runOC("get", "deployment", kueueOperatorDeployment, "-n", kueueOperatorNamespace, "-o", "jsonpath={.spec.template.spec.containers[?(@.name=='"+kueueOperatorDeployment+"')].securityContext.readOnlyRootFilesystem}")
				if err != nil {
					return fmt.Errorf("failed to check read-only root filesystem: %v", err)
				}

				if strings.TrimSpace(readOnlyRootFilesystem) != "true" {
					return fmt.Errorf("root filesystem not read-only")
				}

				return nil
			}, timeout, interval).Should(Succeed())
		})
	})

	Describe("Operator Functionality Test", func() {
		It("should be able to create and manage Kueue instances", func() {
			By("Checking if a Kueue instance can be created")

			// First verify no existing instance
			_, err := runOC("get", "kueue", "cluster")
			if err == nil {
				Skip("Kueue instance already exists, skipping functionality test")
			}

			// This is a basic connectivity test - we're just checking that the operator
			// is running and the API is accessible. A full functional test would require
			// creating actual Kueue instances, but that's beyond the scope of an
			// installation health check.

			By("Verifying the Kueue API is accessible")
			Eventually(func() error {
				// Try to list Kueue instances (should work even if empty)
				_, err := runOC("get", "kueues")
				return err
			}, timeout, interval).Should(Succeed())
		})
	})

	Describe("Operator Recovery Test", func() {
		It("should recover gracefully from pod restart", func() {
			By("Getting initial operator pod names")
			initialPodNames := runOCExpectSuccess("get", "pods", "-n", kueueOperatorNamespace, "-l", "name="+kueueOperatorDeployment, "-o", "jsonpath={.items[*].metadata.name}")
			Expect(strings.TrimSpace(initialPodNames)).NotTo(BeEmpty(), "No initial operator pods found")

			By("Triggering operator pod restart")
			restartTime := time.Now().Format(time.RFC3339)
			runOCExpectSuccess("patch", "deployment", kueueOperatorDeployment, "-n", kueueOperatorNamespace, "-p", fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, restartTime))

			By("Waiting for new operator pods to be ready")
			Eventually(func() error {
				// Get current pod names
				currentPodNames, err := runOC("get", "pods", "-n", kueueOperatorNamespace, "-l", "name="+kueueOperatorDeployment, "-o", "jsonpath={.items[*].metadata.name}")
				if err != nil {
					return fmt.Errorf("failed to get current pods: %v", err)
				}

				if strings.TrimSpace(currentPodNames) == "" {
					return fmt.Errorf("no current pods found")
				}

				currentPods := strings.Fields(currentPodNames)
				initialPods := strings.Fields(initialPodNames)

				// Check if we have new pods (different names) that are ready
				newPods := []string{}
				for _, current := range currentPods {
					isNew := true
					for _, initial := range initialPods {
						if current == initial {
							isNew = false
							break
						}
					}
					if isNew {
						newPods = append(newPods, current)
					}
				}

				if len(newPods) == 0 {
					return fmt.Errorf("no new pods found after restart")
				}

				// Check if new pods are ready
				for _, podName := range newPods {
					ready, err := runOC("get", "pod", podName, "-n", kueueOperatorNamespace, "-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}")
					if err != nil {
						return fmt.Errorf("failed to check pod readiness: %v", err)
					}
					if strings.TrimSpace(ready) != "True" {
						return fmt.Errorf("pod %s not ready", podName)
					}
				}

				return nil
			}, timeout, interval).Should(Succeed())
		})
	})
})
