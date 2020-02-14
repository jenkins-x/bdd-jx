package saas

import (
	"github.com/jenkins-x/bdd-jx/test/helpers"
	"time"

	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	cmd "github.com/jenkins-x/jx/pkg/cmd/clients"
	"github.com/jenkins-x/jx/pkg/kube"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	setupTimeout = 1 * time.Minute
)

type saasConfig struct {
	name          string
	client        kubernetes.Interface
	namespace     string
	baseNamespace string
}

func newSaasConfig(client kubernetes.Interface, namespace, baseNamespace string) *saasConfig {
	return &saasConfig{
		client:        client,
		namespace:     namespace,
		name:          kube.IngressConfigConfigmap,
		baseNamespace: baseNamespace,
	}
}

type testCaseSaas struct {
	*runner.JxRunner
	kubeClient kubernetes.Interface
	namespace  string
}

func newTestCaseSaas(cwd string, factory cmd.Factory, ns string) (*testCaseSaas, error) {
	client, _, err := factory.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	return &testCaseSaas{
		JxRunner:   runner.New(cwd, nil, 0),
		kubeClient: client,
		namespace:  ns,
	}, nil
}

func (t *testCaseSaas) expectIngress(name string) {
	ing, err := t.kubeClient.Extensions().Ingresses(t.namespace).Get(name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(ing.GetName()).To(Equal(name))
}

func (t *testCaseSaas) notExpectIngress(name string) {
	_, err := t.kubeClient.Extensions().Ingresses(t.namespace).Get(name, metav1.GetOptions{})
	Expect(err).To(HaveOccurred())
}

var _ = Describe("upgrade ingress\n", func() {
	var test *testCaseSaas
	BeforeEach(func() {
		var err error
		test, err = newTestCaseSaas(helpers.WorkDir, cmd.NewFactory(), "test-saas")
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
	//AfterEach(func() {
	//	err := test.cleanupCluster()
	//	Expect(err).NotTo(HaveOccurred())
	//})
})
