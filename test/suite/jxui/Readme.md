## Running JXUI test locally
* In CI mode, JXUI test suite will do all pre-requisites for you: setup a cluster and install JXUI app in your cluster.

*  In DEV mode, if you have those pre-requisites setup:
    * [remote cluster installed](https://docs.cloudbees.com/docs/cloudbees-jenkins-x-distribution/latest/install-guide/cluster) 
    * with a local `jxui-frontend` and `jxui-backend` running
you can run JXUI BDD tests suite by pointing to your local JXUI
```
GH_ACCESS_TOKEN=XXXX JXUI_URL=http://localhost:9000 make test-jxui
```