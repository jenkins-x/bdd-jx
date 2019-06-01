package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x/golang-jenkins"
)

func AddCoverageArgsIfNeeded(args []string, id string) ([]string, error){
	if os.Getenv("ENABLE_COVERAGE") == strings.ToLower("true") || os.Getenv("ENABLE_COVERAGE") == strings.ToLower("1") || os.Getenv("ENABLE_COVERAGE") == strings.ToLower("on") {
		reportsDir := os.Getenv("REPORTS_DIR")
		if reportsDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, errors2.Wrapf(err, "getting current dir")
			}
			reportsDir = filepath.Join(cwd, "build","reports")
		}
		outFile := filepath.Join(reportsDir, fmt.Sprintf("%s.coverage.out", id))
		LogInfof("Enabling coverage, writing coverage to %s\n", outFile)
		err := os.Setenv("COVER_JX_BINARY", "true")
		if err != nil {
			return nil, errors2.Wrapf(err, "setting env var COVER_JX_BINARY to true")
		}
		err = os.MkdirAll(reportsDir, 0700)
		if err != nil {
			return nil, errors2.Wrapf(err, "creating coverage dir")
		}
		args = append([]string{"-test.coverprofile", outFile}, args...)
	}
	return args, nil
}

// GetEnv fetches a timeout value from an environment variable, and returns the fallback value if that variable does not exist
func GetTimeoutFromEnv(key string, fallback int) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return time.Duration(intVal)
		}
	}
	return time.Duration(fallback) * time.Minute
}

func GetJenkinsClient() (gojenkins.JenkinsClient, error) {
	url := os.Getenv("BDD_JENKINS_URL")
	if url == "" {
		return nil, errors.New("no BDD_JENKINS_URL env var set. Try running this command first:\n\n  eval $(gofabric8 bdd-env)\n")
	}
	username := os.Getenv("BDD_JENKINS_USERNAME")
	token := os.Getenv("BDD_JENKINS_TOKEN")

	bearerToken := os.Getenv("BDD_JENKINS_BEARER_TOKEN")
	if bearerToken == "" && (token == "" || username == "") {
		return nil, errors.New("no BDD_JENKINS_TOKEN or BDD_JENKINS_BEARER_TOKEN && BDD_JENKINS_USERNAME env var set")
	}

	auth := &gojenkins.Auth{
		Username:    username,
		ApiToken:    token,
		BearerToken: bearerToken,
	}
	jenkins := gojenkins.NewJenkins(auth, url)

	// handle insecure TLS for minishift
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	jenkins.SetHTTPClient(httpClient)
	return jenkins, nil
}

func GetFileAsString(path string) (string, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("No file found at path %s", path)
	}

	return string(buf), nil
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()
	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()
	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}
	return
}

func CopyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	// create dest dir
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}
	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()
		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

func Random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
