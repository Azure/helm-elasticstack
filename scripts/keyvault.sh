cert_folder='certs'

# Handler executed on exit
function on_exit() {
    rm -rf ${cert_folder}
}

# Checks the return code of a command
function check_rc() {
    if [ $? -ne 0 ]
    then
        echo $1
        trap on_exit EXIT
        exit -1
    fi
}

# Fetches a secret from KeyVault
# Arguments:
# $1 - KeyVault name
# $2 - Secret Name
function get_secret() {
    keyvault_name=$1
    secret_name=$2
    local secret=$(az keyvault secret show --vault-name $keyvault_name -n $secret_name --output json --query "value" | tr -d '\"')
    echo $secret
}

# Fetches a PKCS12 certificate from KeyVault and extracts the public certificate and private key
# Arguments:
#  $1 - KeyVault name
#  $2 - Name of the PKCS12 certificate in KeyVault
#  $3 - Domain name for which the certificate was issued
#  $4 - Password with which the private key of the certificate was encrypted in KeyVault
#  $5 - Variable which will hold the public certificate upon function return
#  $6 - variable which will hold the private key upon function return
function get_cert_and_key() {
    keyvault_name=$1
    cert_name=$2
    cert_domain=$3
    key_password=$4
    public_cert=$5
    private_key=$6

    mkdir -p ${cert_folder}

    az keyvault secret download --vault-name ${keyvault_name} -f ${cert_folder}/${cert_domain}.pfx \
       --name ${cert_name} --encoding base64
    check_rc "Failed to fetch from KeyVault the PKCS12 certificate ${cert_name}"

    openssl pkcs12 -in ${cert_folder}/${cert_domain}.pfx -nocerts -out ${cert_folder}/${cert_domain}_key.pem \
            -password pass: -passout pass:${key_password} &> /dev/null
    check_rc "Failed to extract the private key from PKCS12 certificate ${cert_name}"

    openssl rsa -in ${cert_folder}/${cert_domain}_key.pem -out ${cert_folder}/${cert_domain}.key \
            -passin pass:${key_password} &> /dev/null
    check_rc "Failed to convert the private key to RSA key format for certificate ${cert_name}"

    openssl pkcs12 -in ${cert_folder}/${cert_domain}.pfx -clcerts -nokeys -out ${cert_folder}/${cert_domain}.pem \
            -passin pass: &> /dev/null
    check_rc "Failed to extract the public certificate from PKCS12 certificate ${cert_name}"

    public_cert=$(cat ${cert_folder}/${cert_domain}.pem | base64 | tr -d "\n")
    private_key=$(cat ${cert_folder}/${cert_domain}.key | base64 | tr -d "\n")

    rm -rf $cert_folder
}
