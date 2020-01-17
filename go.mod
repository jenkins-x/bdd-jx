module github.com/jenkins-x/bdd-jx

require (
	code.gitea.io/sdk v0.0.0-20180702024448-79a281c4e34a
	github.com/Azure/draft v0.15.0
	github.com/IBM-Cloud/bluemix-go v0.0.0-20181008063305-d718d474c7c2
	github.com/Jeffail/gabs v1.1.1
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e
	github.com/Netflix/go-expect v0.0.0-20180814212900-124a37274874
	github.com/Pallinder/go-randomdata v1.1.0
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/alexflint/go-filemutex v0.0.0-20171028004239-d358565f3c3f
	github.com/andygrunwald/go-gerrit v0.0.0-20181026193842-43cfd7a94eb4
	github.com/andygrunwald/go-jira v1.5.0
	github.com/antham/chyle v1.4.0
	github.com/aws/aws-sdk-go v1.24.0
	github.com/banzaicloud/bank-vaults v0.0.0-20190508130850-5673d28c46bd
	github.com/beevik/etree v1.0.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/bouk/monkey v1.0.0
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/chromedp/cdproto v0.0.0-20180720050708-57cf4773008d
	github.com/chromedp/chromedp v0.1.1
	github.com/codeship/codeship-go v0.0.0-20180717142545-7793ca823354
	github.com/denormal/go-gitignore v0.0.0-20180713143441-75ce8f3e513c
	github.com/fatih/color v1.7.0
	github.com/fatih/structs v1.1.0
	github.com/gfleury/go-bitbucket-v1 v0.0.0-20190216152406-3a732135aa4d
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/jsonreference v0.19.2
	github.com/go-openapi/spec v0.19.2
	github.com/gogo/protobuf v1.2.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.1
	github.com/google/go-cmp v0.3.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/hashicorp/go-version v1.1.0
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/vault v1.1.2
	github.com/heptio/sonobuoy v0.16.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c
	github.com/hpcloud/tail v1.0.0
	github.com/iancoleman/orderedmap v0.0.0-20181121102841-22c6ecc9fe13
	github.com/jenkins-x/draft-repo v0.0.0-20180417100212-2f66cc518135
	github.com/jenkins-x/go-scm v1.5.60
	github.com/jenkins-x/golang-jenkins v0.0.0-20180919102630-65b83ad42314
	github.com/jenkins-x/jx v0.0.0-20200109001856-9433d1c29eca
	github.com/jetstack/cert-manager v0.5.2
	github.com/knative/build v0.7.0
	github.com/knative/build-pipeline v0.1.0
	github.com/knative/pkg v0.0.0-20190624141606-d82505e6c5b4
	github.com/magiconair/properties v1.8.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/nlopes/slack v0.0.0-20180721202243-347a74b1ea30
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pborman/uuid v1.2.0
	github.com/petergtz/pegomock v2.7.0+incompatible
	github.com/pkg/browser v0.0.0-20170505125900-c90ca0c84f15
	github.com/pkg/errors v0.8.1
	github.com/rodaine/hclencoder v0.0.0-20180926060551-0680c4321930
	github.com/russross/blackfriday v1.5.2
	github.com/satori/go.uuid v1.2.1-0.20180103174451-36e9d2ebbde5
	github.com/sethvargo/go-password v0.1.2
	github.com/shirou/gopsutil v0.0.0-20180901134234-eb1f1ab16f2e
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/stoewer/go-strcase v1.0.1
	github.com/stretchr/testify v1.4.0
	github.com/wbrefvem/go-bitbucket v0.0.0-20190128183802-fc08fd046abb
	github.com/xanzy/go-gitlab v0.22.1
	gocloud.dev v0.9.0
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f
	gopkg.in/AlecAivazis/survey.v1 v1.8.3
	gopkg.in/src-d/go-git.v4 v4.5.0
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190718183219-b59d8169aab5
	k8s.io/apiextensions-apiserver v0.0.0-20190718185103-d1ef975d28ce
	k8s.io/apimachinery v0.0.0-20190703205208-4cfb76a8bf76
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20190416052311-01a054e913a9
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a
	k8s.io/helm v2.7.2+incompatible
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/metrics v0.0.0-20180620010437-b11cf31b380b
	k8s.io/test-infra v0.0.0-20190131093439-a22cef183a8f

)

replace github.com/heptio/sonobuoy => github.com/jenkins-x/sonobuoy v0.11.7-0.20190318120422-253758214767

replace k8s.io/api => k8s.io/api v0.0.0-20181128191700-6db15a15d2d3

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20181128195641-3954d62a524d

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7

replace k8s.io/client-go => k8s.io/client-go v2.0.0-alpha.0.0.20190115164855-701b91367003+incompatible

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181128195303-1f84094d7e8e

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/sirupsen/logrus => github.com/jtnord/logrus v1.4.2-0.20190423161236-606ffcaf8f5d

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v21.1.0+incompatible

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v10.15.5+incompatible

replace github.com/banzaicloud/bank-vaults => github.com/banzaicloud/bank-vaults v0.0.0-20190508130850-5673d28c46bd

// TODO: Remove once https://github.com/jenkins-x/go-scm/pull/65 is merged and we can update the real dependency
replace github.com/jenkins-x/go-scm => github.com/abayer/go-scm v1.5.1-0.20200117190616-8db331f9cc60
