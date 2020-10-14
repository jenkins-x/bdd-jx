#!/usr/bin/env bash
set -e

if [ -z "$GIT_USERNAME" ]
then
    export GIT_USERNAME="jenkins-x-labs-bot"
fi

if [ -z "$GIT_SERVER_HOST" ]
then
    export GIT_SERVER_HOST="github.com"
fi

if [ -z "$GH_OWNER" ]
then
    export GH_OWNER="cb-kubecd"
fi

if [ -z "$QUICKSTART" ]
then
    export QUICKSTART="golang"
fi

export GIT_USER_EMAIL="jenkins-x@googlegroups.com"


export GIT_PROVIDER_URL="https://${GIT_SERVER_HOST}"


if [ -z "$GIT_TOKEN" ]
then
      echo "ERROR: no GIT_TOKEN env var defined"
      exit 1
else
      echo "has valid git token"
fi

export GITHUB_TOKEN="${GIT_TOKEN//[[:space:]]}"

export JX_DISABLE_DELETE_APP="true"
export JX_DISABLE_DELETE_REPO="true"

# increase the timeout for complete PipelineActivity
export BDD_TIMEOUT_PIPELINE_ACTIVITY_COMPLETE="60"

# we don't yet update the PipelineActivity.spec.pullTitle on previews....
export BDD_DISABLE_PIPELINEACTIVITY_CHECK="true"

# define variables for the BDD tests
export GIT_ORGANISATION="$GH_OWNER"
export GH_USERNAME="$GIT_USERNAME"

echo "Running the BDD tests for $QUICKSTART"

./build/bddjx -ginkgo.focus=$QUICKSTART -test.v
