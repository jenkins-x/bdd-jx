package saas

import (
	"fmt"
	"github.com/jenkins-x/bdd-jx/test/helpers"
	cmd "github.com/jenkins-x/jx/pkg/cmd/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type testCaseSaas struct {
	kubeClient kubernetes.Interface
	namespace  string
}

func newTestCaseSaas(cwd string, factory cmd.Factory, ns string) (*testCaseSaas, error) {
	client, _, err := factory.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	return &testCaseSaas{
		kubeClient: client,
		namespace:  ns,
	}, nil
}

func (t *testCaseSaas) expectIngress(name string) {
	ing, err := t.kubeClient.ExtensionsV1beta1().Ingresses(t.namespace).Get(name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(ing.GetName()).To(Equal(name))
}

func (t *testCaseSaas) expectPod(name string, count int) {
	listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", name)}
	pods, err := t.kubeClient.CoreV1().Pods(t.namespace).List(listOptions)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(pods.Items)).To(Equal(count))
}

func (t *testCaseSaas) expectAllPodsNotInState(phase v1.PodPhase) {
	listOptions := metav1.ListOptions{}
	pods, err := t.kubeClient.CoreV1().Pods(t.namespace).List(listOptions)
	Expect(err).NotTo(HaveOccurred())
	for _, pod := range pods.Items {
		if !strings.Contains(pod.Labels["job-name"],"jx-boot") {
			Expect(pod.Status.Phase).NotTo(Equal(phase), fmt.Sprintf("pod %s is in phase %s", pod.Name, pod.Status.Phase))
		}
	}
}

func (t *testCaseSaas) notExpectIngress(name string) {
	_, err := t.kubeClient.ExtensionsV1beta1().Ingresses(t.namespace).Get(name, metav1.GetOptions{})
	Expect(err).To(HaveOccurred())
}

var _ = Describe("SaaS Configuration\n", func() {
	var test *testCaseSaas
	BeforeEach(func() {
		var err error
		test, err = newTestCaseSaas(helpers.WorkDir, cmd.NewFactory(), "jx")
		Expect(err).NotTo(HaveOccurred())
		Expect(test).NotTo(BeNil())
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("vault does not have an ingress\n", func() {
				const testSvc = "vault"
				test.notExpectIngress(testSvc)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("nexus does not have an ingress\n", func() {
				const testSvc = "nexus"
				test.notExpectIngress(testSvc)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("chartmuseum does not have an ingress\n", func() {
				const testSvc = "chartmuseum"
				test.notExpectIngress(testSvc)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("hook does have an ingress\n", func() {
				const testSvc = "hook"
				test.expectIngress(testSvc)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("kuberhealthy does have an ingress\n", func() {
				const testSvc = "kuberhealthy"
				test.expectIngress(testSvc)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("jenkins-x-bucketrepo pod is running\n", func() {
				const testPod = "jenkins-x-bucketrepo"
				test.expectPod(testPod, 1)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("kuberhealthy pod is running\n", func() {
				const testPod = "kuberhealthy"
				test.expectPod(testPod, 2)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("jenkins-x-jx-segment-controller pod is running\n", func() {
				const testPod = "jenkins-x-jx-segment-controller"
				test.expectPod(testPod, 1)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("jenkins-x-repositorycontroller pod is running\n", func() {
				const testPod = "jenkins-x-repositorycontroller"
				test.expectPod(testPod, 1)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("tekton-pipelines-controller pod is running\n", func() {
				const testPod = "tekton-pipelines-controller"
				test.expectPod(testPod, 1)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("crier pod is not running\n", func() {
				const testPod = "crier"
				test.expectPod(testPod, 0)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("hook pod is not running\n", func() {
				const testPod = "hook"
				test.expectPod(testPod, 0)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("jenkins-x-nexus pod is not running\n", func() {
				const testPod = "jenkins-x-nexus"
				test.expectPod(testPod, 0)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("tide pod is running\n", func() {
				const testPod = "tide"
				test.expectPod(testPod, 1)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when a saas cluster is configured", func() {
			It("pods not in Failed|Unknown state\n", func() {
				test.expectAllPodsNotInState(v1.PodFailed)
				test.expectAllPodsNotInState(v1.PodUnknown)
			})
		})
	})
	//AfterEach(func() {
	//	err := test.cleanupCluster()
	//	Expect(err).NotTo(HaveOccurred())
	//})
})
