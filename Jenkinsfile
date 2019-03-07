pipeline {
    agent any

    environment {
        JENKINS_CREDS       = credentials('test-jenkins-user')
        GH_CREDS            = credentials('jx-pipeline-git-github-github')
        GHE_CREDS           = credentials('jx-pipeline-git-github-ghe')
        GKE_SA              = credentials('gke-sa')

        GIT_USERNAME        = "$GH_CREDS_USR"
        GIT_API_TOKEN       = "$GH_CREDS_PSW"
        GITHUB_ACCESS_TOKEN = "$GH_CREDS_PSW"

        JOB_NAME            = "$JOB_NAME"
        BRANCH_NAME         = "$BRANCH_NAME"
        ORG                 = 'jenkinsxio'

        // Build and tests configuration (run only 2 builds/tests in parallel
        // in order to avoid OOM issue
        PARALLEL_BUILDS = 1

        // BDD tests configuration
        GIT_PROVIDER_URL     = "https://github.beescloud.com"
        GHE_TOKEN            = "$GHE_CREDS_PSW"
        GINKGO_ARGS          = "-v"

        JX_DISABLE_DELETE_APP  = "true"
        JX_DISABLE_DELETE_REPO = "true"
    }
    options {
        skipDefaultCheckout(true)
    }
    stages {
        stage('CI Build and Test') {
            when {
                anyOf {
                    environment name: 'JOB_TYPE', value: 'presubmit'
                    environment name: 'JOB_TYPE', value: 'postsubmit'
                    environment name: 'JOB_TYPE', value: 'batch'
                }
            }
            steps {
                dir ('/workspace') {
                    // lets create a team for this PR and run the BDD tests
                    sh "gcloud auth activate-service-account --key-file $GKE_SA"
                    sh "gcloud container clusters get-credentials anthorse --zone europe-west1-b --project jenkinsx-dev"

                    // image doesn't have an up-to-date jx version (it's missing the --skip-test-git-repo-clone flag)
                    sh "wget https://github.com/jenkins-x/jx/releases/download/v1.3.875/jx-linux-amd64.tar.gz"
                    sh "tar -xzf jx-linux-amd64.tar.gz"
                    sh "cp jx /usr/bin"

                    // lets setup git
                    sh "git config --global --add user.name JenkinsXBot"
                    sh "git config --global --add user.email jenkins-x@googlegroups.com"

                    // Copy the current code into /home/jenkins/go/jenkins-x/bdd-jx as it needs to be there to run the tests
                    sh "mkdir -p /home/jenkins/go/jenkins-x/bdd-jx"
                    sh "cp -r * /home/jenkins/go/jenkins-x/bdd-jx"

                    // Run the BDD tests using the currently checked-out code
                    sh "jx step bdd -b --skip-test-git-repo-clone --provider=gke --git-provider=ghe --git-provider-url=https://github.beescloud.com --git-username dev1 --git-api-token $GHE_CREDS_PSW --default-admin-password $JENKINS_CREDS_PSW --no-delete-app --no-delete-repo --tests install --tests test-parallel"
                }
            }
        }
    }
}