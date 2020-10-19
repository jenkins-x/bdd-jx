package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

var activityLineRegex = regexp.MustCompile(`(?m:(^.*?)\s*((?:\d+h)?(?:\d+m)?(?:\d+s))?\s*((?:\d+h)?(?:\d+m)?(?:\d+s))\s*(.*)$)`)

type Activity struct {
	JobName     string
	BuildNumber int
	StartedAgo  string
	Duration    string
	Status      string
	Stages      []*Stage
}

type Stage struct {
	Name       string
	StartedAgo string
	Duration   string
	Status     string
	Steps      []*Step
}

type Step struct {
	Name       string
	StartedAgo string
	Duration   string
	Status     string
}

func ParseJxGetActivities(s string) (map[string]*Activity, error) {
	answer := make(map[string]*Activity, 0)
	lines := strings.Split(strings.TrimSpace(s), "\n")
	// Trim the header row
	var currentActivity *Activity
	var currentStage *Stage
	headerFound := false
	defaultJobName := ""
	for _, line := range lines {
		// Ignore any output before the header
		if strings.HasPrefix(line, "STEP") {
			headerFound = true
			continue
		}
		if !headerFound {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			// If the string starts with text, it's the root of an activity
			line = strings.TrimSpace(line)
			fields := activityLineRegex.FindStringSubmatch(line)
			if len(fields) != 5 {
				if defaultJobName == "" {
					defaultJobName = line
				}
				fmt.Printf("ignoring activity output line: %s\n", line)
				continue
			}
			currentActivity = &Activity{
				JobName:    fields[1],
				StartedAgo: fields[2],
				Duration:   fields[3],
				Status:     fields[4],
				Stages:     make([]*Stage, 0),
			}
			answer[currentActivity.JobName] = currentActivity
		} else if !strings.HasPrefix(line, "    ") {
			line = strings.TrimSpace(line)
			fields := activityLineRegex.FindStringSubmatch(line)
			if len(fields) == 5 {
				currentStage = &Stage{
					Name:       fields[1],
					StartedAgo: fields[2],
					Duration:   fields[3],
					Status:     fields[4],
				}
			} else {
				currentStage = &Stage{
					Name: line,
				}
			}
			if currentActivity == nil {
				currentActivity = &Activity{
					JobName: defaultJobName,
				}
				answer[currentActivity.JobName] = currentActivity
			}
			currentActivity.Stages = append(currentActivity.Stages, currentStage)
		} else {
			// If the regex matches, it's a Step
			line = strings.TrimSpace(line)
			fields := activityLineRegex.FindStringSubmatch(line)
			if len(fields) == 5 {
				step := &Step{
					Name:       fields[1],
					StartedAgo: fields[2],
					Duration:   fields[3],
					Status:     fields[4],
				}
				currentStage.Steps = append(currentStage.Steps, step)
			} else {
				fmt.Printf("Ignoring output line: %s", line)
			}
		}
	}
	if currentActivity != nil && currentActivity.Status == "" && len(currentActivity.Stages) > 0 {
		currentActivity.Status = currentActivity.Stages[0].Status
	}
	return answer, nil
}
