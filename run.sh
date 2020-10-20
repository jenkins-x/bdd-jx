#!/usr/bin/env bash
set -e

echo "starting the Jenkins X BDD tests..."

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

# lets turn off color output
export TERM=dumb

export JX_DISABLE_DELETE_APP="true"
export JX_DISABLE_DELETE_REPO="true"

# increase the timeout for complete PipelineActivity
export BDD_TIMEOUT_PIPELINE_ACTIVITY_COMPLETE="60"

# we don't yet update the PipelineActivity.spec.pullTitle on previews....
export BDD_DISABLE_PIPELINEACTIVITY_CHECK="true"

# define variables for the BDD tests
export GIT_ORGANISATION="$GH_OWNER"
export GH_USERNAME="$GIT_USERNAME"

# lets ensure that git is setup
export CURRENT_GIT_USER_NAME=$(git config --global --get user.name)
export CURRENT_GIT_USER_EMAIL=$(git config --global --get user.email)

if [ -z "$CURRENT_GIT_USER_NAME" ]
then
    git config --global --add user.name ${GIT_USERNAME:-jenkins-x-bot}
fi
if [ -z "$CURRENT_GIT_USER_EMAIL" ]
then
    git config --global --add user.email ${GIT_EMAIL:-jenkins-x@googlegroups.com}
fi

export CURRENT_GIT_USER_NAME=$(git config --global --get user.name)
export CURRENT_GIT_USER_EMAIL=$(git config --global --get user.email)

echo "git user name:  $CURRENT_GIT_USER_NAME"
echo "git user email: $CURRENT_GIT_USER_EMAIL"

echo "Running the BDD tests for $QUICKSTART"

bddjx -ginkgo.focus=$QUICKSTART -test.v
