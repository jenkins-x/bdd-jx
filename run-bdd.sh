#!/bin/sh

export GHE_CREDS_PSW=dedc04db4212426f90aa0b3cbc9b56a2532497de
export JENKINS_CREDS_PSW=jxadmin18
export GIT_PROVIDER_URL=https://github.beescloud.com

export GHE_TOKEN=$GHE_CREDS_PSW

export GINKGO_ARGS=-v

make $*
