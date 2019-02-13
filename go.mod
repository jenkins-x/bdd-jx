module github.com/jenkins-x/bdd-jx

require (
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/google/pprof v0.0.0-20190109223431-e84dfd68c163 // indirect
	github.com/jenkins-x/golang-jenkins v0.0.0-20180919102630-65b83ad42314
	github.com/jenkins-x/jx v1.3.871
	github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega v1.4.3
	github.com/pkg/errors v0.8.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/spf13/viper v1.3.1 // indirect
	golang.org/x/arch v0.0.0-20181203225421-5a4828bb7045 // indirect
	golang.org/x/net v0.0.0-20190119204137-ed066c81e75e // indirect
	gopkg.in/src-d/go-git.v4 v4.8.1
	k8s.io/api v0.0.0-20190126160303-ccdd560a045f
	k8s.io/apiextensions-apiserver v0.0.0-20181128195303-1f84094d7e8e
	k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/metrics v0.0.0-20180620010437-b11cf31b380b
)

replace github.com/heptio/sonobuoy => github.com/jenkins-x/sonobuoy v0.11.7-0.20190131193045-dad27c12bf17

replace k8s.io/api => k8s.io/api v0.0.0-20181128191700-6db15a15d2d3

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20181128195641-3954d62a524d

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7

replace k8s.io/client-go => k8s.io/client-go v2.0.0-alpha.0.0.20190115164855-701b91367003+incompatible
