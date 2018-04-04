#!/bin/bash
# Copyright (c) Microsoft and contributors.  All rights reserved.
#
# This source code is licensed under the MIT license found in the
# LICENSE file in the root directory of this source tree.

# Include some common functions
current_dir="$(dirname $0)"
source "$current_dir/../../scripts/util.sh"
source "$current_dir/../../scripts/keyvault.sh"

function show_help() {
    cat <<EOF
Usage: ${0##*/} [-h] [-t] [-e ENVIRONMENT] [-n NAMESPACE] [-l LICENSE_FILE]
Deploys a Kubernetes Helm chart with in a given environment and namespace.
         -h               display this help and exit
         -e ENVIRONMENT   environment for which the deployment is perfomed (e.g. acs)
         -n NAMESPACE     namespace where the cluster will be deployed
         -l LICENSE_FILE  license file for x-pack elasticsearch
         -r RELEASE_NAME  Helm release name
         -t               validate only the templates without performing any deployment
EOF
}

CHART_NAME="elasticsearch"
RELEASE_NAME="elasticsearch"
ENVIRONMENT='acs'
NAMESPACE='elk'
LICENSE=''
DRY_RUN=false

while getopts he:l:tn:r: opt; do
    case $opt in
        h)
            show_help
            exit 0
            ;;
        e)
            ENVIRONMENT=$OPTARG
            ;;
        l)
            LICENSE=$OPTARG
            ;;
        t)
            DRY_RUN=true
            ;;
        n)
            NAMESPACE=$OPTARG
            ;;
        r)
            RELEASE_NAME=$OPTARG
            ;;
        *)
            show_help >&2
            exit 1
            ;;
    esac
done

helm_values=" -f environments/${ENVIRONMENT}/values.yaml"
helm_params=""

# Check if the required command are installed
echo "Checking kubectl command"
type kubectl > /dev/null 2>&1
check_rc "kubectl command not found in \$PATH. Please follow the documentation to install it: https://kubernetes.io/docs/tasks/kubectl/install/"

echo "Checking helm command"
type helm > /dev/null 2>&1
check_rc "helm command not found in \$PATH. Please follow the documentation to install it: https://github.com/kubernetes/helm"

# Read the Elasticsearch x-pack license file if defined
if [ -e "$LICENSE" ]; then
    license=$(cat $LICENSE | base64 | tr -d "\n")
    helm_params+=" --set license.value=${license}"
else
    echo "Installing Elasticsearch cluster without x-pack license"
fi

# Install or upgrades the helm chart
echo -n "Installing $CHART_NAME helm chart..."
error=$(mktemp)
output=$(mktemp)
(

    if [[ "$DRY_RUN" = true ]]
    then
        helm template $helm_values $helm_params .
    else
        helm upgrade -i --reset-values --timeout 1800 --namespace $NAMESPACE $helm_values $helm_params $RELEASE_NAME . --wait &> $output
    fi

    if [ $? -ne 1 ]
    then
        echo " OK" > $error
    else
        echo " FAIL" > $error
    fi
) &
spinner
if [ $(cat $error) == " FAIL" ]
then
    echo "Fail"
    cat $output
    exit -1
fi

if [[ "$DRY_RUN" = true ]]
then
    echo " Done"
    cat $output
else
    echo " Done"
fi
