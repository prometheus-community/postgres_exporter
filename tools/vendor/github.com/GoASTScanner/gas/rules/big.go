// (c) Copyright 2016 Hewlett Packard Enterprise Development LP
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rules

import (
	"go/ast"

	"github.com/GoASTScanner/gas"
)

type usingBigExp struct {
	gas.MetaData
	pkg   string
	calls []string
}

func (r *usingBigExp) Match(n ast.Node, c *gas.Context) (gi *gas.Issue, err error) {
	if _, matched := gas.MatchCallByType(n, c, r.pkg, r.calls...); matched {
		return gas.NewIssue(c, n, r.What, r.Severity, r.Confidence), nil
	}
	return nil, nil
}

// NewUsingBigExp detects issues with modulus == 0 for Bignum
func NewUsingBigExp(conf gas.Config) (gas.Rule, []ast.Node) {
	return &usingBigExp{
		pkg:   "*math/big.Int",
		calls: []string{"Exp"},
		MetaData: gas.MetaData{
			What:       "Use of math/big.Int.Exp function should be audited for modulus == 0",
			Severity:   gas.Low,
			Confidence: gas.High,
		},
	}, []ast.Node{(*ast.CallExpr)(nil)}
}
