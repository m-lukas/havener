// Copyright © 2018 The Havener
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package havener

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// ProcessConfigFile reads the havener config file from the provided path and
// returns the processed input file. Any shell operator will be evaluated.
func ProcessConfigFile(path string) (*Config, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = yaml.Unmarshal(source, &config); err != nil {
		return nil, err
	}

	for idx, release := range config.Releases {
		if release.Overrides == nil {
			continue
		}

		config.Releases[idx].Overrides, err = TraverseStructureAndProcessShellOperators(release.Overrides)
		if err != nil {
			return nil, err
		}
	}

	return &config, nil
}

// TraverseStructureAndProcessShellOperators traverses the provided generic structure
// and processes all string leafs.
func TraverseStructureAndProcessShellOperators(input interface{}) (interface{}, error) {
	var err error

	switch obj := input.(type) {
	case map[interface{}]interface{}:
		for key, value := range obj {
			obj[key], err = TraverseStructureAndProcessShellOperators(value)
			if err != nil {
				return nil, err
			}
		}

	case []interface{}:
		for idx, value := range obj {
			obj[idx], err = TraverseStructureAndProcessShellOperators(value)
			if err != nil {
				return nil, err
			}
		}

	case string:
		input, err = processShellOperator(obj)
		if err != nil {
			return nil, err
		}

	case nil:
		input, err = map[interface{}]interface{}{}, nil
	}

	return input, err
}

// processShellOperator processes the input string and evaluates any shell
// operator in it.
func processShellOperator(s string) (string, error) {
	// https://regex101.com/r/dvdiTp/2
	shellRegexp := regexp.MustCompile(`\({2}\s*shell\s*(.+?)\s*\B\){2}`)
	if matches := shellRegexp.FindAllStringSubmatch(s, -1); len(matches) > 0 {
		for _, match := range matches {
			/* #0 is the whole string,
			 * #1 is the command
			 */
			cmd := exec.Command("sh", "-c", match[1])

			var out bytes.Buffer
			cmd.Stdout = &out

			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to run command: %s\nerror message: %s", match[1], err.Error())
			}

			result := strings.TrimSpace(out.String())
			s = strings.Replace(s, match[0], result, -1)
		}
	}

	return s, nil
}
