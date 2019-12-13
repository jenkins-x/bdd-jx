#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [[ -z "${REPORTS_DIR}" ]]; then
  DEFAULT_TEST_REPORT_DIR=${DIR}/../../build/reports/
  echo "Did not find report dir, defaulting to ${DEFAULT_TEST_REPORT_DIR}"
  TEST_REPORTS_DIR=${DEFAULT_TEST_REPORT_DIR}
else
  echo "Found report dir: ${REPORTS_DIR}"
  TEST_REPORTS_DIR="${REPORTS_DIR}"
fi

npm install
echo "Running cypress tests with APPLICATION_NAME=$APPLICATION_NAME"
CYPRESS_APPLICATION_NAME=$APPLICATION_NAME npm test
TEST_EXIT_CODE=$?

echo "Moving test report to ${TEST_REPORTS_DIR}"
mv ui-smoke.junit.xml "${TEST_REPORTS_DIR}" || true

exit ${TEST_EXIT_CODE}