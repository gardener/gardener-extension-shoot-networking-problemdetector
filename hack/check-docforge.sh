#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

docCommitHash="fa2e9f84851be81e85668986675db235bb43a6b5"

echo "> Check Docforge Manifest"
repoPath=${1-"$(readlink -f "$(dirname "${0}")/..")"}
manifestPath=${2-"${repoPath}/.docforge/manifest.yaml"}
diffDirs=${3-".docforge/;docs/"}
repoName=${4-"gardener"}
useToken=${5-false}

tmpDir=$(mktemp -d)
function cleanup {
    rm -rf "$tmpDir"
}
trap cleanup EXIT ERR INT TERM

curl https://raw.githubusercontent.com/gardener/documentation/${docCommitHash}/.ci/check-manifest --output "${tmpDir}/check-manifest-script.sh" && chmod +x "${tmpDir}/check-manifest-script.sh"
# fix main branch
sed -i 's|origin/master|origin/main|g' "${tmpDir}/check-manifest-script.sh"
curl https://raw.githubusercontent.com/gardener/documentation/${docCommitHash}/.ci/check-manifest-config --output "${tmpDir}/manifest-config"
scriptPath="${tmpDir}/check-manifest-script.sh"
configPath="${tmpDir}/manifest-config"

${scriptPath} --repo-path "${repoPath}" --repo-name "${repoName}" --use-token "${useToken}" --manifest-path "${manifestPath}" --diff-dirs "${diffDirs}" --config-path "${configPath}"