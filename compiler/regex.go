/*
 * Copyright © 2016-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
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
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright 	2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 */

// Package compiler offers a regexp compiler which compiles regex templates to regexp.Regexp
//
//  reg, err := compiler.CompileRegex("foo:bar.baz:<[0-9]{2,10}>", '<', '>')
//  // if err != nil ...
//  reg.MatchString("foo:bar.baz:123")
//
//  reg, err := compiler.CompileRegex("/foo/bar/url/{[a-z]+}", '{', '}')
//  // if err != nil ...
//  reg.MatchString("/foo/bar/url/abz")
//
// This package is adapts github.com/gorilla/mux/regexp.go

package compiler

// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license as follows:

// Copyright (c) 2012 Rodrigo Moraes. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
// * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
// * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/dlclark/regexp2"
)

const regexp2MatchTimeout = time.Millisecond * 250

// delimiterIndices returns the first level delimiter indices from a string.
// It returns an error in case of unbalanced delimiters.
func delimiterIndices(s string, delimiterStart, delimiterEnd byte) ([]int, error) {
	var level, idx int
	idxs := make([]int, 0)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case delimiterStart:
			if level++; level == 1 {
				idx = i
			}
		case delimiterEnd:
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf(`Unbalanced braces in "%q"`, s)
			}
		}
	}

	if level != 0 {
		return nil, fmt.Errorf(`Unbalanced braces in "%q"`, s)
	}

	return idxs, nil
}

// CompileRegex parses a template and returns a Regexp.
//
// You can define your own delimiters. It is e.g. common to use curly braces {} but I recommend using characters
// which have no special meaning in Regex, e.g.: <, >
//
//  reg, err := compiler.CompileRegex("foo:bar.baz:<[0-9]{2,10}>", '<', '>')
//  // if err != nil ...
//  reg.MatchString("foo:bar.baz:123")
func CompileRegex(tpl string, delimiterStart, delimiterEnd byte) (*regexp2.Regexp, error) {
	// Check if it is well-formed.
	idxs, errBraces := delimiterIndices(tpl, delimiterStart, delimiterEnd)
	if errBraces != nil {
		return nil, errBraces
	}
	varsR := make([]*regexp2.Regexp, len(idxs)/2)
	pattern := bytes.NewBufferString("")
	pattern.WriteByte('^')

	var end int
	for i := 0; i < len(idxs); i += 2 {
		// Set all values we are interested in.
		raw := tpl[end:idxs[i]]
		end = idxs[i+1]
		patt := tpl[idxs[i]+1 : end-1]
		// Build the regexp pattern.
		varIdx := i / 2
		fmt.Fprintf(pattern, "%s(%s)", regexp.QuoteMeta(raw), patt)
		reg, err := regexp2.Compile(fmt.Sprintf("^%s$", patt), regexp2.RE2)
		if err != nil {
			return nil, err
		}
		reg.MatchTimeout = regexp2MatchTimeout
		varsR[varIdx] = reg
	}

	// Add the remaining.
	raw := tpl[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	pattern.WriteByte('$')

	// Compile full regexp.
	reg, errCompile := regexp2.Compile(pattern.String(), regexp2.RE2)
	if errCompile != nil {
		return nil, errCompile
	}
	reg.MatchTimeout = regexp2MatchTimeout

	return reg, nil
}
