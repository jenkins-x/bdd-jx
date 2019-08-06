package parsers





import (
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

var activityLineRegex = regexp.MustCompile(`(?m:(^.*?)\s*((?:\d+h)?(?:\d+m)?(?:\d+s))?\s*((?:\d+h)?(?:\d+m)?(?:\d+s))\s*(.*)$)`)

type Activity struct {
	JobName string
	BuildNumber int
	StartedAgo string
	Duration string
	Status string
	Stages []*Stage
}

type Stage struct {
	Name string
	StartedAgo string
	Duration string
	Status string
	Steps []*Step
}

type Step struct {
	Name string
	StartedAgo string
	Duration string
	Status string
}

func ParseJxGetActivities(s string) (map[string]*Activity, error) {
	answer := make(map[string]*Activity, 0)
	lines := strings.Split(strings.TrimSpace(s), "\n")
	// Trim the header row
	var currentActivity *Activity
	var currentStage *Stage
	headerFound := false
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
			if len(fields) != 5{
				return nil, errors.Errorf("unable to parse %s as activity, entire output was: \n\n%s\n", line, s)
			}
			currentActivity = &Activity{
				JobName:fields[1],
				StartedAgo: fields[2],
				Duration: fields[3],
				Status: fields[4],
				Stages: make([]*Stage, 0),
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
			} else if fields[2] == "Pending" {
				// This means that the step hasn't started yet.
				step := &Step{
					Name: fields[1],
					Status: fields[2],
				}
				currentStage.Steps = append(currentStage.Steps, step)
			} else {
				return  nil, errors.Errorf("Unable to parse %s as step, entire output was %s", line, s)
			}
		}
	}
	return answer, nil
}
