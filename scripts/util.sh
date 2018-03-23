# Shows a spinner
function spinner() {
    local pid=$!
    local delay=0.75
    local spinstr='|/-\'
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Checks the return code of an command
function check_return_code() {
    if [ $? -ne 0 ]
    then
        echo $1
        trap on_exit EXIT
        exit -1
    fi
}
