#!/bin/bash

# This script detects non-compliant licenses in the output of language-specific license checkers.

# These licenses are allowed.
# These are the exact and complete license strings for 100% legal certainty, no regexes.
ALLOWED_LICENSES=(
	'0BSD'
	'AFLv2.1'
	'AFLv2.1,BSD'
	'(AFL-2.1 OR BSD-3-Clause)'
	'Apache 2.0'
	'Apache-2.0'
	'(Apache-2.0 OR MPL-1.1)'
	'Apache-2.0 AND MIT'
	'Apache License, Version 2.0'
	'Apache*'
	'Artistic-2.0'
	'BlueOak-1.0.0'
	'BSD'
	'BSD*'
	'BSD-2-Clause'
	'(BSD-2-Clause OR MIT OR Apache-2.0)'
	'BSD-3-Clause'
	'(BSD-3-Clause OR GPL-2.0)'
	'BSD-3-Clause OR MIT'
	'CC0-1.0'
	'CC-BY-3.0'
	'CC-BY-4.0'
	'(CC-BY-4.0 AND MIT)'
	'ISC'
	'ISC*'
	'LGPL-2.1' # LGPL allows commercial use, requires only that modifications to LGPL-protected libraries are published under a GPL-compatible license
	'MIT'
	'MIT*'
	'MIT-0'
	'MIT AND ISC'
	'(MIT AND BSD-3-Clause)'
	'(MIT AND Zlib)'
	'(MIT OR Apache-2.0)'
	'(MIT OR CC0-1.0)'
	'(MIT OR GPL-2.0)'
	'MPL-2.0'
	'(MPL-2.0 OR Apache-2.0)'
	'Public Domain'
	'Python-2.0' # the Python-2.0 is a permissive license, see https://en.wikipedia.org/wiki/Python_License
	'Unlicense'
	'WTFPL'
	'WTFPL OR ISC'
	'(WTFPL OR MIT)'
	'(MIT OR WTFPL)'
	'LGPL-3.0-or-later' # Requires only that modifications to LGPL-protected libraries are published under a GPL-compatible license which is not the case at Ory
)

# These modules don't work with the current license checkers
# and have been manually verified to have a compatible license (regex format).
APPROVED_MODULES=(
	'https://github.com/ory-corp/cloud/'                                                  # Ory IP
	'github.com/ory/hydra-client-go'                                                      # Apache-2.0
	'github.com/ory/hydra-client-go/v2'                                                   # Apache-2.0
	'github.com/ory/kratos-client-go'                                                     # Apache-2.0
	'github.com/gobuffalo/github_flavored_markdown'                                       # MIT
	'buffers@0.1.1'                                                                       # MIT: original source at http://github.com/substack/node-bufferlist is deleted but a fork at https://github.com/pkrumins/node-bufferlist/blob/master/LICENSE contains the original license by the original author (James Halliday)
	'https://github.com/iconify/iconify/packages/react'                                   # MIT: license is in root of monorepo at https://github.com/iconify/iconify/blob/main/license.txt
	'github.com/gobuffalo/.*'                                                             # MIT: license is in root of monorepo at https://github.com/gobuffalo/github_flavored_markdown/blob/main/LICENSE
	'github.com/ory-corp/cloud/.*'                                                        # Ory IP
	'github.com/golang/freetype/.*'                                                       # FreeType license: https://freetype.sourceforge.net/FTL.TXT
	'go.opentelemetry.io/otel/exporters/jaeger/internal/third_party/thrift/lib/go/thrift' # Incorrect detection, actually Apache-2.0: https://github.com/open-telemetry/opentelemetry-go/blob/exporters/jaeger/v1.17.0/exporters/jaeger/internal/third_party/thrift/LICENSE
	'go.uber.org/zap/exp/.*'                                                              # MIT license is in root of exp folder in monorepo at https://github.com/uber-go/zap/blob/master/exp/LICENSE
	'github.com/ory/client-go'                                                            # Apache-2.0
	'github.com/ian-kent/linkio'                                                          # BSD - https://github.com/ian-kent/linkio/blob/97566b8728870dac1c9863ba5b0f237c39166879/linkio.go#L1-L3
	'github.com/t-k/fluent-logger-golang/fluent'                                          # Apache-2.0 https://github.com/t-k/fluent-logger-golang/blob/master/LICENSE
	'github.com/jmespath/go-jmespath'                                                     # Apache-2.0 https://github.com/jmespath/go-jmespath/blob/master/LICENSE
	'github.com/ory/keto/proto/ory/keto/opl/v1alpha1'                                     # Apache-2.0 - submodule of keto
	'github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2'                         # Apache-2.0 - submodule of keto
)

# These lines in the output should be ignored (plain text, no regex).
IGNORE_LINES=(
	'"module name","licenses"' # header of license output for Node.js
)

echo_green() {
	printf "\e[1;92m%s\e[0m\n" "$@"
}

echo_red() {
	printf "\e[0;91m%s\e[0m\n" "$@"
}

# capture STDIN
input=$(cat -)

# remove ignored lines
for ignored in "${IGNORE_LINES[@]}"; do
	input=$(echo "$input" | grep -vF "$ignored")
done

# remove pre-approved modules
for approved in "${APPROVED_MODULES[@]}"; do
	input=$(echo "$input" | grep -v "\"${approved}\"")
	input=$(echo "$input" | grep -v "\"Custom: ${approved}\"")
done

# remove allowed licenses
for allowed in "${ALLOWED_LICENSES[@]}"; do
	input=$(echo "$input" | grep -vF "\"${allowed}\"")
done

# anything left in the input at this point is a module with an invalid license

# print outcome
if [ -z "$input" ]; then
	echo_green "Licenses are okay."
else
	echo_red "Unknown licenses found!"
	echo "$input"
	exit 1
fi
