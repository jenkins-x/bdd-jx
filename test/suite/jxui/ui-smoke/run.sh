#!/bin/sh

set -e

[ "$APPLICATION_NAME" ] || (echo 'ERROR: APPLICATION_NAME env var not set' >&2; false)
[ "$JENKINS_PASSWORD" ] || (echo 'ERROR: JENKINS_PASSWORD env var not set' >&2; false)

DIR="$(cd "$(dirname "$0")" && pwd)" || exit 1
UI_PROXY_PORT="$(node get-free-port.js)" || exit 1

startProxy() {
  echo Starting local proxy...
  PROXY_PORT="$UI_PROXY_PORT" \
  TARGET_BASE_URL="$CYPRESS_BASE_URL" \
    npm run proxy &
  export CYPRESS_BASE_URL="http://127.0.0.1:$UI_PROXY_PORT"
  for i in $(seq 1 5); do
    curl -fsS "http://127.0.0.1:$UI_PROXY_PORT" >/dev/null && return 0
    sleep 1
  done
  echo 'Timed out waiting for proxy to start' >&2
  return 1
}

runCypressTests() {
  echo "Running cypress tests with APPLICATION_NAME=$APPLICATION_NAME on CYPRESS_BASE_URL=$CYPRESS_BASE_URL"
  CYPRESS_APPLICATION_NAME=$APPLICATION_NAME npm test
}

moveTestReports() {
  if [ -z "${REPORTS_DIR}" ]; then
    DEFAULT_TEST_REPORT_DIR=${DIR}/../../build/reports
    echo "Did not find report dir, defaulting to ${DEFAULT_TEST_REPORT_DIR}"
    TEST_REPORTS_DIR=${DEFAULT_TEST_REPORT_DIR}
  else
    echo "Found report dir: ${REPORTS_DIR}"
    TEST_REPORTS_DIR="${REPORTS_DIR}"
  fi
  echo "Moving test report to ${TEST_REPORTS_DIR}"
  mv ui-smoke.junit.xml "${TEST_REPORTS_DIR}" || true

  echo "Moving screenshots to ${TEST_REPORTS_DIR}/screenshots"
  rm -rf "${TEST_REPORTS_DIR}/screenshots" || true
  mv "./cypress/screenshots" "${TEST_REPORTS_DIR}" || true
}

npm install
startProxy
runCypressTests || TEST_EXIT_CODE=$?
moveTestReports
curl "http://127.0.0.1:$UI_PROXY_PORT/shutdown-proxy" --connect-timeout 3 2>/dev/null || true
exit ${TEST_EXIT_CODE}
