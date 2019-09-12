package devpods

import (
	"strings"
	"time"

	"github.com/jenkins-x/bdd-jx/test/helpers"
	"github.com/jenkins-x/bdd-jx/test/utils"
	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	cmd "github.com/jenkins-x/jx/pkg/jx/cmd/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type TestDevPods struct {
	*runner.JxRunner
	kubeClient kubernetes.Interface
}

func newTestDevPods(factory cmd.Factory) (*TestDevPods, error) {

	timeOut := utils.GetTimeoutFromEnv("BDD_TIMEOUT_DEVPOD", 15)

	client, _, err := factory.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	return &TestDevPods{
		JxRunner:   runner.New(helpers.WorkDir, &timeOut, 0),
		kubeClient: client,
	}, nil
}

func getAllAvailableDevpods() []string {
	factory := cmd.NewFactory()
	kc, _, err := factory.CreateKubeClient()
	if err != nil {
		utils.LogInfof("Error creating kube client %s", err)
	}

	configMapInterface := kc.CoreV1().ConfigMaps("jx")

	list, err := configMapInterface.List(metav1.ListOptions{
		LabelSelector: "jenkins.io/kind=podTemplate",
	})
	if err != nil {
		utils.LogInfof("Error querying config map %s", err)
	}

	devPods := make([]string, 0)
	for _, cm := range list.Items {

		devPod := strings.TrimPrefix(cm.Name, "jenkins-x-pod-template-")

		if isDevPod(devPod) && !isExcluded(devPod) {
			utils.LogInfof("Adding devpod %s \n", devPod)
			devPods = append(devPods, devPod)
		}
	}
	utils.LogInfof("Added all devPods %s", devPods)
	//return []string{"go", "maven"}
	return devPods
}

func isDevPod(label string) bool {
	//these options currently error
	return label != "terraform" && label != "packer" && label != "jx-base" && label != "promote" && label != "swift" && label != "ruby"
}

func isExcluded(label string) bool {
	//machine learning ones take extremely long time to provision
	return strings.Contains(label, "machine-learning")
}

func (test *TestDevPods) createDevPod(label string) {
	args := []string{"create", "devpod", "-b", "-l", label, "--import=false", "--suffix=devpod"}
	test.Run(args...)
}

func (test *TestDevPods) getPodName(suffix string) string {
	listOptions := metav1.ListOptions{}

	pods, err := test.kubeClient.CoreV1().Pods("jx").List(listOptions)

	Expect(err).NotTo(HaveOccurred())

	var name string

	for _, pod := range pods.Items {
		if strings.HasSuffix(pod.Name, suffix) {
			name = pod.Name
		}
	}
	Expect(name).ShouldNot(Equal(""))
	return name
}

func (test *TestDevPods) checkDevPodExists(label string) {
	name := label + "-devpod"
	utils.LogInfof("Creating dev pod %s", label)
	args := []string{"get", "devpod"}
	devPods, err := test.RunWithOutput(args...)
	utils.ExpectNoError(err)

	Expect(devPods).Should(ContainSubstring(name))
}

func (test *TestDevPods) checkDevPodTerminating() {
	//jx get devpod
	args := []string{"get", "devpod"}

	var b = false
	var i = 0

	//wait up to two mintutes to terminate
	for b == false && i < 120 {
		devPods, err := test.RunWithOutput(args...)
		utils.ExpectNoError(err)
		b = !strings.Contains(devPods, "Terminating")
		time.Sleep(1 * time.Second)
		i++
	}
}

func (test *TestDevPods) checkDevPodNoLongerExists(label string) {
	name := label + "-devpod"
	//jx get devpod
	utils.LogInfof("checking pod no longer exists %s", label)
	args := []string{"get", "devpod"}
	devPods, err := test.RunWithOutput(args...)
	utils.ExpectNoError(err)
	Expect(devPods).ShouldNot(ContainSubstring(name))
}

func (test *TestDevPods) deleteDevPods(label string) {
	name := label + "-devpod"
	utils.LogInfof("deleting pod %s", label)
	//jx delete devpod
	podName := test.getPodName(name)
	args := []string{"delete", "devpod", podName, "-b"}
	test.Run(args...)
}

var _ = Describe("E2E tests for all Dev pods \n", func() {

	// will get all available devpods
	devPods := getAllAvailableDevpods()

	// and create a set of tests for each of them
	for _, label := range devPods {
		var _ = Describe("Given I have created a devpod with label "+label, func() {
			// copy the label, otherwise by the time the closure executes it's value will of changed
			devpod := string([]byte(label))
			var test *TestDevPods
			BeforeEach(func() {
				var err error
				test, err = newTestDevPods(cmd.NewFactory())

				Expect(err).NotTo(HaveOccurred())
				Expect(test).NotTo(BeNil())
			})
			Context("when running jx create devpod -l "+label, func() {
				It("a "+label+" dev pod is created\n", func() {
					test.createDevPod(devpod)
				})
			})
			Context("when checking if a devpod exists", func() {
				It("the dev pod is available ", func() {
					test.checkDevPodExists(devpod)
				})
			})
			Context("when running jx delete devpod ", func() {
				It("the dev pod is delete ", func() {
					test.deleteDevPods(devpod)
					test.checkDevPodTerminating()
				})

			})
			Context("when checking if the devpod exists ", func() {
				It("the devpod is no longer available ", func() {
					test.checkDevPodNoLongerExists(devpod)
				})
			})
		})
	}

})
