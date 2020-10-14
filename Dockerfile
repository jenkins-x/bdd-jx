FROM gcr.io/jenkinsxio/jx-cli-base:0.0.30

COPY ./build/linux/bddjx /usr/bin/bddjx
COPY run.sh /usr/bin/runbddjx.sh

ENTRYPOINT /usr/bin/runbddjx.sh