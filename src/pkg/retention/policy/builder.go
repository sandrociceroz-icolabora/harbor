// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg/or"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors"
	"github.com/pkg/errors"
)

// Builder builds the runnable processor from the raw policy
type Builder interface {
	// Builds runnable processor
	//
	//  Arguments:
	//    policy *LiteMeta : the simple metadata of retention policy
	//
	//  Returns:
	//    Processor : a processor implementation to process the candidates
	//    error     : common error object if any errors occurred
	Build(policy *LiteMeta) (alg.Processor, error)
}

// NewBuilder news a basic builder
func NewBuilder(all []*res.Candidate) Builder {
	return &basicBuilder{
		allCandidates: all,
	}
}

// basicBuilder is default implementation of Builder interface
type basicBuilder struct {
	allCandidates []*res.Candidate
}

// Build policy processor from the raw policy
func (bb *basicBuilder) Build(policy *LiteMeta) (alg.Processor, error) {
	if policy == nil {
		return nil, errors.New("nil policy to build processor")
	}

	switch policy.Algorithm {
	case AlgorithmOR:
		// New OR processor
		p := or.New()
		for _, r := range policy.Rules {
			evaluator, err := rule.Get(r.Template, r.Parameters)
			if err != nil {
				return nil, err
			}

			perf, err := action.Get(r.Action, bb.allCandidates)
			if err != nil {
				return nil, errors.Wrap(err, "get action performer by metadata")
			}

			sl := make([]res.Selector, 0)
			for _, s := range r.TagSelectors {
				sel, err := selectors.Get(s.Kind, s.Decoration, s.Pattern)
				if err != nil {
					return nil, errors.Wrap(err, "get selector by metadata")
				}

				sl = append(sl, sel)
			}

			p.AddEvaluator(evaluator, sl)
			p.AddActionPerformer(r.Action, perf)

			return p, nil
		}
	default:
	}

	return nil, errors.Errorf("algorithm %s is not supported", policy.Algorithm)
}