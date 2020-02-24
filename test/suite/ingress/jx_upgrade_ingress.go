package ingress

import (
	"fmt"
	"github.com/jenkins-x/bdd-jx/test/utils"
	"strconv"
	"time"

	"github.com/jenkins-x/bdd-jx/test/helpers"

	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	cmd "github.com/jenkins-x/jx/pkg/cmd/clients"
	"github.com/jenkins-x/jx/pkg/kube"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

const (
	setupTimeout = 1 * time.Minute
)

type ingressConfig struct {
	name          string
	client        kubernetes.Interface
	namespace     string
	baseNamespace string
	config        *kube.IngressConfig
}

func newIngressConfig(client kubernetes.Interface, namespace, baseNamespace string) *ingressConfig {
	return &ingressConfig{
		client:        client,
		namespace:     namespace,
		name:          kube.IngressConfigConfigmap,
		baseNamespace: baseNamespace,
		config:        &kube.IngressConfig{},
	}
}

func (ic *ingressConfig) toConfigMapData() map[string]string {
	return map[string]string{
		kube.Domain:  ic.config.Domain,
		kube.Email:   ic.config.Email,
		kube.Exposer: ic.config.Exposer,
		kube.Issuer:  ic.config.Issuer,
		kube.TLS:     strconv.FormatBool(ic.config.TLS),
	}
}

func (ic *ingressConfig) fromConfigMapData(data map[string]string) {
	if data == nil {
		return
	}
	domain, ok := data[kube.Domain]
	if ok {
		ic.config.Domain = domain
	}
	email, ok := data[kube.Email]
	if ok {
		ic.config.Email = email
	}
	expser, ok := data[kube.Exposer]
	if ok {
		ic.config.Exposer = expser
	}
	issuer, ok := data[kube.Issuer]
	if ok {
		ic.config.Issuer = issuer
	}
	tls, ok := data[kube.TLS]
	if ok {
		tls, err := strconv.ParseBool(tls)
		if err == nil {
			ic.config.TLS = tls
		}
	}
}
func (ic *ingressConfig) set(config *kube.IngressConfig) {
	ic.config = config
}

func (ic *ingressConfig) merge(config *kube.IngressConfig) {
	if config.Domain != "" {
		ic.config.Domain = config.Domain
	}
	if config.Email != "" {
		ic.config.Email = config.Email
	}
	if config.Exposer != "" {
		ic.config.Exposer = config.Exposer
	}
	if config.Issuer != "" {
		ic.config.Issuer = config.Issuer
	}
	ic.config.TLS = config.TLS
}

func (ic *ingressConfig) create() error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ic.name,
			Namespace: ic.namespace,
		},
		Data: ic.toConfigMapData(),
	}
	_, err := ic.client.CoreV1().ConfigMaps(ic.namespace).Create(cm)
	return err
}

func (ic *ingressConfig) update() error {
	cm, err := ic.client.CoreV1().ConfigMaps(ic.namespace).Get(ic.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "retrieving the current ingress config")
	}
	cm.Data = ic.toConfigMapData()
	_, err = ic.client.CoreV1().ConfigMaps(ic.namespace).Update(cm)
	if err != nil {
		return errors.Wrap(err, "updating the ingress config")
	}
	return nil
}

func (ic *ingressConfig) updateBase() error {
	cm, err := ic.client.CoreV1().ConfigMaps(ic.baseNamespace).Get(ic.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "retrieving the current ingress config from base namespace")
	}
	cm.Data = ic.toConfigMapData()
	_, err = ic.client.CoreV1().ConfigMaps(ic.baseNamespace).Update(cm)
	if err != nil {
		return errors.Wrap(err, "updating the ingress config into the base namespace")
	}
	return nil
}

func (ic *ingressConfig) copyFromBaseNamespace() error {
	cm, err := ic.client.CoreV1().ConfigMaps(ic.baseNamespace).Get(ic.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "copying config form %s/%s", ic.baseNamespace, ic.name)
	}
	ic.fromConfigMapData(cm.Data)
	return nil
}

type testCaseUpgradeIngress struct {
	*runner.JxRunner
	kubeClient kubernetes.Interface
	namespace  string
	ic         *ingressConfig
}

func newTestCaseUpgradeIngress(cwd string, factory cmd.Factory, ns string) (*testCaseUpgradeIngress, error) {
	client, _, err := factory.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	return &testCaseUpgradeIngress{
		JxRunner:   runner.New(cwd, nil, 0),
		kubeClient: client,
		namespace:  ns,
		ic:         newIngressConfig(client, ns, "jx"),
	}, nil
}

func (t *testCaseUpgradeIngress) setupCluster() error {
	// ensure that the test namespace does not already exists
	err := utils.Retry(setupTimeout, func() error {
		_, err := t.kubeClient.CoreV1().Namespaces().Get(t.namespace, metav1.GetOptions{})
		if err != nil {
			return nil
		}
		return fmt.Errorf("test namespace %q exists", t.namespace)
	})
	if err != nil {
		return errors.Wrapf(err, "checking tests namespace %q does not exist", t.namespace)
	}

	// create the test namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.namespace,
		},
	}
	_, err = t.kubeClient.CoreV1().Namespaces().Create(ns)
	if err != nil {
		return errors.Wrapf(err, "creating the test namespace %q", t.namespace)
	}

	// create the ingress configuration
	err = t.ic.copyFromBaseNamespace()
	if err != nil {
		return errors.Wrap(err, "copying the ingress config from base namespace")
	}
	err = t.ic.create()
	if err != nil {
		return errors.Wrapf(err, "creating the ingress config in namespace %q", t.namespace)
	}

	return nil
}

func (t *testCaseUpgradeIngress) cleanupCluster() error {
	err := t.kubeClient.CoreV1().Namespaces().Delete(t.namespace, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "deleting the test namesapce %q", t.namespace)
	}
	return nil
}

func (t *testCaseUpgradeIngress) createService(name string) error {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"fabric8.io/expose": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "http",
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
	_, err := t.kubeClient.CoreV1().Services(t.namespace).Create(svc)
	if err != nil {
		return errors.Wrapf(err, "creating service %s/%s", t.namespace, name)
	}
	return nil
}

func (t *testCaseUpgradeIngress) expectIngress(name string) {
	ing, err := t.kubeClient.Extensions().Ingresses(t.namespace).Get(name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(ing.GetName()).To(Equal(name))
}

func (t *testCaseUpgradeIngress) notExpectIngress(name string) {
	_, err := t.kubeClient.Extensions().Ingresses(t.namespace).Get(name, metav1.GetOptions{})
	Expect(err).To(HaveOccurred())
}

var _ = Describe("upgrade ingress\n", func() {
	var test *testCaseUpgradeIngress
	BeforeEach(func() {
		var err error
		test, err = newTestCaseUpgradeIngress(helpers.WorkDir, cmd.NewFactory(), "test-upgrade-ingress")
		Expect(err).NotTo(HaveOccurred())
		Expect(test).NotTo(BeNil())

		err = test.setupCluster()
		Expect(err).NotTo(HaveOccurred())
	})
	Describe("Given valid parameters", func() {
		Context("when running upgrade ingress", func() {
			It("creates the ingress resource for services from specified namespace\n", func() {
				const testSvc = "test-svc"
				err := test.createService(testSvc)
				Expect(err).NotTo(HaveOccurred())

				test.Run("upgrade", "ingress", "-b", "--skip-resources-update",
					"--namespaces="+test.namespace, "--config-namespace="+test.namespace)

				test.expectIngress(testSvc)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when running upgrade ingress", func() {
			It("creates the ingress resource only for selected service from specified namespace\n", func() {
				const testSvc1 = "test-svc1"
				const testSvc2 = "test-svc2"
				err := test.createService(testSvc1)
				Expect(err).NotTo(HaveOccurred())
				err = test.createService(testSvc2)
				Expect(err).NotTo(HaveOccurred())

				test.Run("upgrade", "ingress", "-b", "--skip-resources-update",
					"--services="+testSvc1,
					"--namespaces="+test.namespace, "--config-namespace="+test.namespace)

				test.expectIngress(testSvc1)
				test.notExpectIngress(testSvc2)
			})
		})
	})
	Describe("Given valid parameters", func() {
		Context("when running upgrade ingress", func() {
			It("creates the ingress resource and fetch a TLS certificate for services from specified namespace\n", func() {
				const testSvc = "test-svc"
				err := test.createService(testSvc)
				Expect(err).NotTo(HaveOccurred())

				originalConfig := *test.ic.config
				test.ic.merge(&kube.IngressConfig{
					Issuer: "letsencrypt-staging",
					Email:  "test@jenkinsx.io",
					TLS:    true,
				})
				err = test.ic.update()
				Expect(err).NotTo(HaveOccurred())

				test.Run("upgrade", "ingress", "-b", "--skip-resources-update",
					"--namespaces="+test.namespace, "--config-namespace="+test.namespace,
					"--wait-for-certs")

				test.expectIngress(testSvc)

				test.ic.set(&originalConfig)
				err = test.ic.updateBase()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
	AfterEach(func() {
		err := test.cleanupCluster()
		Expect(err).NotTo(HaveOccurred())
	})
})
