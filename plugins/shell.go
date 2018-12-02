package plugins

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ghetzel/go-stockutil/log"
)

type ShellPlugin struct {
	Plugin
	ExecPath         []string
	PerPluginTimeout time.Duration
	MaxTimeout       time.Duration
}

func (self ShellPlugin) autotype(value string) interface{} {
	rxBoolTrue := regexp.MustCompile(`(?i)^(?:true|t|1|on|yes|enabled?|active)$`)
	rxBoolFalse := regexp.MustCompile(`(?i)^(?:false|f|0|off|no|disabled?|inactive)$`)
	rxFloat := regexp.MustCompile(`^-?[0-9]+\.[0-9]+$`)
	rxInteger := regexp.MustCompile(`^-?[0-9]+$`)
	rxDate3339 := regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[\-\+][0-9]{2}:?[0-9]{2}$`)

	if rxBoolTrue.MatchString(value) {
		return true
	} else if rxBoolFalse.MatchString(value) {
		return false
	} else if rxFloat.MatchString(value) {
		rv, _ := strconv.ParseFloat(value, 32)
		return rv
	} else if rxInteger.MatchString(value) {
		rv, _ := strconv.ParseInt(value, 10, 64)
		return rv
	} else if rxDate3339.MatchString(value) {
		if rv, err := time.Parse(time.RFC3339, value); err == nil {
			return rv
		} else {
			return value
		}
	} else {
		return value
	}
}

func (self ShellPlugin) totype(value string, typename string) (interface{}, error) {
	rxBoolTrue := regexp.MustCompile(`(?i)^(?:true|t|1|on|yes|enabled?|active)$`)
	rxBoolFalse := regexp.MustCompile(`(?i)^(?:false|f|0|off|no|disabled?|inactive)$`)

	if typename == "bool" && rxBoolTrue.MatchString(value) {
		return true, nil
	} else if typename == "bool" && rxBoolFalse.MatchString(value) {
		return false, nil
	} else if typename == "float" {
		rv, err := strconv.ParseFloat(value, 32)
		return rv, err
	} else if typename == "int" {
		rv, err := strconv.ParseInt(value, 10, 64)
		return rv, err
	} else if typename == "date" {
		rv, err := time.Parse(time.RFC3339, value)
		return rv, err
	} else if typename == "str" {
		return value, nil
	}

	return nil, fmt.Errorf("Invalid type '%s'", typename)
}

func (self ShellPlugin) executePluginCommand(cmdPath string, values chan Observation, waiter *sync.WaitGroup) {
	log.Debugf("Executing %s", cmdPath)
	defer waiter.Done()

	pluginCmd := exec.Command(cmdPath)
	observations := make([]Observation, 0)

	pOut, _ := pluginCmd.StdoutPipe()
	pErr, _ := pluginCmd.StderrPipe()

	done := make(chan bool)

	//  command execution goroutine
	go func() {
		err := pluginCmd.Start()

		if err != nil {
			log.Errorf("Unable to execute command '%s': %s", cmdPath, err)
			waiter.Done()
			return
		}

		// stderr processing
		go func() {
			scanner := bufio.NewScanner(pErr)
			for scanner.Scan() {
				line := scanner.Text()
				log.Errorf("%s: %s", cmdPath, line)
			}
		}()

		// stdout processing
		scanner := bufio.NewScanner(pOut)

		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.SplitN(line, ":", 3)

			nameRx := regexp.MustCompile(`^[0-9a-zA-Z\.\_\-]+$`)
			typeRx := regexp.MustCompile(`^(?:bool|int|float|date|str)?$`)

			//  an informational fact in three parts
			//  Each line of standard output from a plugin script must conform to this spec:
			//
			//  fact.name:type:value\n
			//
			//  fact name: any alphanumeric ASCII character, dot (.), dash (-), or underscore (_)
			//  type:      any one of 'bool', 'int', 'float', 'str', or an empty string; empty will assume automatic type detection
			//  value:     any valid value for the explicitly-stated type, or anything at all for implicit (automatic) types
			//             e.g.: '8' will become an int64
			//                   'false' will become a bool
			//                   'Hooray!' will become a string
			//
			if len(fields) == 3 {
				//validate field name
				if nameRx.MatchString(fields[0]) != true {
					log.Warningf("Invalid field name '%s'", fields[0])
					continue
				}

				//validate type (if specified)
				if typeRx.MatchString(fields[1]) != true {
					log.Warningf("Invalid type '%s'", fields[1])
					continue
				}

				field := fields[0]
				typename := fields[1]
				value := strings.TrimSpace(fields[2])

				//blank type means auto type
				if fields[1] == "" {
					observations = append(observations, Observation{
						Name:  field,
						Value: self.autotype(value),
					})
				} else {
					if v, err := self.totype(value, typename); err == nil {
						observations = append(observations, Observation{
							Name:  field,
							Value: v,
						})
					} else {
						log.Errorf("Field %s: error converting '%s' to type %s: %s", field, value, typename, err)
					}
				}

			} else {
				log.Warningf("Invalid input line '%s'", line)
				continue
			}
		}

		if err := pluginCmd.Wait(); err != nil {
			log.Errorf("%s: %s", cmdPath, err)
		} else {
			for _, observation := range observations {
				values <- observation
			}

			log.Debugf("Command %s finished, got %d observations", cmdPath, len(observations))
		}

		done <- true
	}()

	select {
	case <-done:
		return
	case <-time.After(self.PerPluginTimeout):
		log.Warningf("Timed out waiting for plugin '%s' to complete", cmdPath)

		if pluginCmd.Process != nil {
			log.Debugf("Killing subprocess %d", pluginCmd.Process.Pid)

			if err := pluginCmd.Process.Kill(); err != nil {
				log.Debugf("Failed to kill subprocess %d: %v", pluginCmd.Process.Pid, err)
			}
		}
	}
}

func (self ShellPlugin) closeAfterCommandsFinish(values chan Observation, waiter *sync.WaitGroup) {
	done := make(chan bool)

	go func() {
		waiter.Wait()
		done <- true
	}()

	select {
	case <-done:
		break
	case <-time.After(self.MaxTimeout):
		log.Warning("Timed out waiting for all plugins to complete")
	}

	close(values)
}

func (self ShellPlugin) Collect() ([]Observation, error) {
	var waiter sync.WaitGroup
	values := make([]Observation, 0)
	valChan := make(chan Observation)

	for _, p := range self.ExecPath {
		if strings.HasPrefix(p, `~`) {
			if u, err := user.Current(); err == nil {
				p = strings.Replace(p, `~`, u.HomeDir, -1)
			} else {
				log.Warningf("Failed to determine current user while expanding path %q", p)
			}
		}

		files := make([]string, 0)

		// handle platform-specific implementations
		platformRoot := path.Join(p, fmt.Sprintf("platform-%s", runtime.GOOS))
		platformFiles, _ := ioutil.ReadDir(platformRoot)
		for _, file := range platformFiles {
			//is the file a file and is it executable?
			if !file.IsDir() && file.Mode()&0111 != 0 {
				files = append(files, path.Join(platformRoot, file.Name()))
			}
		}

		// handle global implementations
		globalFiles, _ := ioutil.ReadDir(p)
		for _, file := range globalFiles {
			//is the file a file and is it executable?
			if !file.IsDir() && file.Mode()&0111 != 0 {
				files = append(files, path.Join(p, file.Name()))
			}
		}

		// execute all paths
		for _, fullPath := range files {
			log.Debugf("Calling exec for %s", fullPath)
			waiter.Add(1)
			go self.executePluginCommand(fullPath, valChan, &waiter)
		}
	}

	go self.closeAfterCommandsFinish(valChan, &waiter)

	for observation := range valChan {
		values = append(values, observation)
	}

	return values, nil
}
