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

package ladon

// Metric is used to expose metrics about authz
type Metric interface {
	// RequestDeniedBy is called when we get explicit deny by policy
	RequestDeniedBy(Request, Policy)
	// RequestAllowedBy is called when a matching policy has been found.
	RequestAllowedBy(Request, Policies)
	// RequestNoMatch is called when no policy has matched our request
	RequestNoMatch(Request)
	// RequestProcessingError is called when unexpected error occured
	RequestProcessingError(Request, Policy, error)
}
