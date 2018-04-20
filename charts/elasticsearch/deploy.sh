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
Usage: ${0##*/} [-h] [-t] [-e ENVIRONMENT] [-n NAMESPACE] [-v VAULT_NAME]
Deploys a Kubernetes Helm chart with in a given environment and namespace.
         -h               display this help and exit
         -e ENVIRONMENT   environment for which the deployment is perfomed (e.g. acs)
         -n NAMESPACE     namespace where the cluster will be deployed
         -r RELEASE_NAME  Helm release name
         -t               validate only the templates without performing any deployment
         -v VAULT_NAME    name of the Aure KeyVault
EOF
}

CHART_NAME="elasticsearch"
RELEASE_NAME="elasticsearch"
ENVIRONMENT='acs'
NAMESPACE='elk'
LICENSE=''
KEYVAULT_NAME=''
ELASTICSEARCH_LICENSE_SECRET='elasticsearch-license'
DRY_RUN=false

while getopts he:l:tn:r:v: opt; do
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
        v)
            KEYVAULT_NAME=$OPTARG
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

# Retrieve the Elasticsearch x-pack license if a KeyVault is provided
if [[ "$DRY_RUN" = false ]]
then
    echo "Checking az command"
    type az > /dev/null 2>&1
    check_rc "az command not found in \$PATH. Please follow the documentation to install it: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"

    echo "Checking openssl command"
    type openssl > /dev/null 2>&1
    check_rc "openssl command not found in \$PATH. Please install it and run again this script."

    if [[ -z "$KEYVAULT_NAME" ]]
    then
        echo "Installing Elasticsearch cluster without x-pack license"
    else
        echo "Retrieving secrets from Azure KeyVault:"
        echo "  Fetching x-pack license"
        license=$(get_secret $KEYVAULT_NAME $ELASTICSEARCH_LICENSE_SECRET)
        check_rc "Failed to fetch the x-pack license from KeyVault"
        helm_params+=" --set license.value=${license}"
    fi
fi

# Install or upgrades the helm chart
echo -n "Installing $CHART_NAME helm chart..."
error=$(mktemp)
output=$(mktemp)
(
    if [[ "$DRY_RUN" = true ]]
    then
        helm template --namespace $NAMESPACE $helm_values $helm_params .
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
