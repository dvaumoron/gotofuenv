/*
 *
 * Copyright 2024 gotofuenv authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package tofuversion

import (
	"fmt"

	"github.com/dvaumoron/gotofuenv/config"
	"github.com/dvaumoron/gotofuenv/tofuversion/tfparser"
	"github.com/hashicorp/go-version"
)

func alwaysTrue(string) bool {
	return true
}

func cmpVersion(v1Str string, v2Str string) int {
	v1, err1 := version.NewVersion(v1Str)
	v2, err2 := version.NewVersion(v2Str)

	if hasErr1, hasErr2 := err1 != nil, err2 != nil; hasErr1 {
		if hasErr2 {
			return 0
		}
		return -1
	} else if hasErr2 {
		return 1
	}
	return v1.Compare(v2)
}

// the boolean second returned value indicates reverse ordering for filtering
func parsePredicate(requestedVersion string, conf *config.Config) (func(string) bool, bool, error) {
	predicate := alwaysTrue
	reverseOrder := true
	switch requestedVersion {
	case config.MinRequiredKey:
		reverseOrder = false // start with older
		fallthrough          // same predicate retrieving
	case config.LatestAllowedKey:
		requireds, err := tfparser.GatherRequiredVersion(conf)
		if err != nil {
			return nil, false, err
		}

		var constraint version.Constraints
		for _, required := range requireds {
			temp, err := version.NewConstraint(required)
			if err != nil {
				return nil, false, err
			}
			constraint = append(constraint, temp...)
		}
		if len(constraint) == 0 {
			reverseOrder = true // erase min-required case
			if conf.Verbose {
				fmt.Println("No OpenTofu version requirement found in files, fallback to latest")
			}
		} else {
			predicate = predicateFromConstraint(constraint)
		}
	case config.LatestKey:
		// nothing to do (alwaysTrue and reverseOrder will work)
	default:
		constraint, err := version.NewConstraint(requestedVersion)
		if err != nil {
			return nil, false, err
		}
		predicate = predicateFromConstraint(constraint)
	}
	return predicate, reverseOrder, nil
}

func predicateFromConstraint(constraint version.Constraints) func(string) bool {
	return func(versionStr string) bool {
		v, err := version.NewVersion(versionStr)
		if err != nil {
			return false
		}
		return constraint.Check(v)
	}
}
