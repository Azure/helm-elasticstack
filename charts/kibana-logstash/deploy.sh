#!/bin/bash
# Copyright (c) Microsoft and contributors.  All rights reserved.
#
# This source code is licensed under the MIT license found in the
# LICENSE file in the root directory of this source tree.

# Include some common functions
current_dir="$(dirname $0)"
source "$current_dir/../../scripts/util.sh"
source "$current_dir/../../scripts/keyvault.sh"


# Configures the Helm parameters which are environment specific
# Arguments:
# $1 - environment name
function get_environment_params() {
    environment=$1
    params=""
    case $environment in
        'acs')
            for env in "${!redis_key_secrets_acs[@]}"; do
                logstash_redis_key=$(get_secret ${redis_key_secrets_acs[$env]})
                check_return_code "Failed to fetch from KeyVault the Redis Key for '$env' environment"
                params+=" --set stunnel.connections.$env.redis.key=${logstash_redis_key}"
            done
            ;;
        *)
            echo "Environment '$environment' not supported!"
            exit 1
            ;;
    esac

    echo $params
}

function show_help() {
    cat <<EOF
Usage: ${0##*/} [-h] [-t] [-e ENVIRONMENT] -d DOMAIN -v VAULT_NAME
Deploys a Kubernetes Helm chart with in a given environment and namespace.
         -h               display this help and exit
         -e ENVIRONMENT   environment for which the deployment is perfomed (e.g. acs)
         -d DOMAIN        public domain name used by the $CHART_NAME
         -n NAMESPACE     namespace where the chart will be deployed
         -v VAULT_NAME    name of the Aure KeyVault where all the secretes and certificates are stored
         -t               validate only the templates without performing any deployment (dry run)
EOF
}

# Predefined constants
CHART_NAME='kibana-logstash'
ENVIRONMENT='acs'
DOMAIN=''
KEYVAULT_NAME=''
NAMESPACE='elk'
DRY_RUN=false

# Predefined KeyVault secrets names
KIBANA_CERTIFICATE_SECRET='kibana-certificate'
KIBANA_CERTIFICATE_KEY_PASSWORD_SECRET='kibana-certificate-key-password'
KIBANA_OAUTH_COOKIE_SECRET='kibana-oauth-cookie-secret'
KIBANA_OAUTH_CLIENT_ID='kibana-oauth-client-id'
KIBANA_OAUTH_CLIENT_SECRET='kibana-oauth-client-secret'
ELASTICSEARCH_WATCHER_WEBHOOK_TEAMS='elasticsearch-watcher-webhook-teams'

# acs environment specific secrets
declare -A redis_key_secrets_acs
redis_key_secrets_acs['dev']='logstash-dev-redis-key'

while getopts hd:e:tn:v: opt; do
    case $opt in
        h)
            show_help
            exit 0
            ;;
        d)
            DOMAIN=$OPTARG
            ;;
        e)
            ENVIRONMENT=$OPTARG
            ;;
        t)
            DRY_RUN=true
            ;;
        n)
            NAMESPACE=$OPTARG
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

# Check if the required commands are installed
echo "Checking helm command"
type helm > /dev/null 2>&1
check_rc "helm command not found in \$PATH. Please follow the documentation to install it: https://github.com/kubernetes/helm"

echo "Checking kubectl command"
type kubectl > /dev/null 2>&1
check_rc "kubectl command not found in \$PATH. Please follow the documentation to install it: https://kubernetes.io/docs/tasks/kubectl/install/"

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
        echo "Please provide the Azure KeyVault where the secrets and certificates are stored!"
        show_help
        exit -1
    fi

    # Set the domain name used by Kibana
    if [[ -z "$DOMAIN" ]]
    then
        echo "Please provide the public domain used by $CHART_NAME!"
        show_help
        exit -1
    fi
    helm_params+=" --set kibana.ingress.host=${DOMAIN}"

    echo "Retrieving secrets from Azure KeyVault:"

    # Fetch form KeyVault the Kibana certificate and private key
    echo "  Fetching Kibana certificate and private key"
    kibana_cert_key_password=$(get_secret $KIBANA_CERTIFICATE_KEY_PASSWORD_SECRET)
    check_rc "Failed to fetch from KeyVault the password for Kibana certificate key"
    get_cert_and_key public_cert private_key $KIBANA_CERTIFICATE_SECRET $kibana_cert_key_password
    helm_params+=" --set kibana.ingress.public.cert=${public_cert}"
    helm_params+=" --set kibana.ingress.private.key=${private_key}"

    # Fetch from KeyVault the OAuth2 proxy secrets
    echo "  Fetching OAuth2 proxy secrets"
    kibana_oauth_client_id=$(get_secret $KIBANA_OAUTH_CLIENT_ID)
    check_rc "Failed to fetch from KeyVault the Kibana OAuth2 Client ID"
    helm_params+=" --set oauth.client.id=${kibana_oauth_client_id}"
    kibana_oauth_client_secret=$(get_secret $KIBANA_OAUTH_CLIENT_SECRET)
    check_rc "Failed to fetch from KeyVault the Kibana OAuth2 Client Secret"
    helm_params+=" --set oauth.client.secret=${kibana_oauth_client_secret}"
    kibana_oauth_cookie_secret=$(get_secret $KIBANA_OAUTH_COOKIE_SECRET)
    check_rc "Failed to fetch from KeyVault the Kibana OAuth2 Cookie Secret"
    helm_params+=" --set oauth.cookie.secret=${kibana_oauth_cookie_secret}"

    # Fetch from KeyVault the Watcher secrets
    echo "  Fetching Elasticsearch Watcher secrets"
    elasticsearch_watcher_webhook_teams=$(get_secret $ELASTICSEARCH_WATCHER_WEBHOOK_TEAMS)
    check_rc "Failed to fetch from KeyVault the Elasticsearch Watcher webhook teams"
    helm_params+=" --set watcher.webhooks.teams=${elasticsearch_watcher_webhook_teams}"

    # Get the environment specific parameters
    echo "  Fetching '$ENVIRONMENT' environemnt specific secrets"
    helm_params+=" $(get_environment_params $ENVIRONMENT)"
fi

# Installing helm chart
echo "Installing $CHART_NAME helm chart"
error=$(mktemp)
output=$(mktemp)
(
    if [[ "$DRY_RUN" = true ]]
    then
        helm template $helm_values $helm_params .
    else
        helm upgrade -i --timeout 1800 --namespace $NAMESPACE $helm_values $helm_params $CHART_NAME . --wait &> $output
    fi

    if [ $? -ne 1 ]
    then
        echo "OK" > $error
    else
        echo "FAIL" > $error
    fi
) &
spinner
if [ $(cat $error) == "FAIL" ]
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
