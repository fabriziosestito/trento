#!/bin/bash

set -e

readonly ARGS="$@"
readonly PROGNAME="./install-on-k3s"
readonly K3S_VERSION="v1.21.3+k3s1"

usage() {
    cat <<- EOF
    usage: $PROGNAME options

    Install Trento server on K3S

    OPTIONS:
       -t --image-tag           image 
       -p --private-key             private key
       -h --help                show this help


    Examples:
       Run all tests:
       $PROGNAME --image 

       Run specific test:
       $PROGNAME --test test_string.sh

       Run:
       $PROGNAME --config /path/to/config/$PROGNAME.conf

       Just show what you are going to do:
       $PROGNAME -vn -c /path/to/config/$PROGNAME.conf
EOF
}


cmdline() {
    local arg=
    for arg
    do
        local delim=""
        case "$arg" in
            #translate --gnu-long-options to -g (short options)
            --image)        args="${args}-i ";;
            --private-key)  args="${args}-p ";;
            --help)         args="${args}-h ";;
            #pass through anything else
            *) [[ "${arg:0:1}" == "-" ]] || delim="\""
            args="${args}${delim}${arg}${delim} ";;
        esac
    done
    
    # Reset the positional parameters to the short options
    eval set -- $args
    
    while getopts "iph" OPTION
    do
        case $OPTION in
            h)
                usage
                exit 0
            ;;
            i)
                readonly IMAGE=$OPTARG
            ;;
            p)
                readonly PRIVATE_KEY=$OPTARG
            ;;
        esac
    done
    
    # if [[ $recursive_testing || -z $RUN_TESTS ]]; then
    #     [[ ! -f $PRIVATE_KEY ]] \
    #     && exit "You must provide --config file"
    # fi
    return 0
}



install_k3s() {
    echo "Installing k3s..."

    curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644
}

install_helm() {
    echo "Installing helm..."
}

update_helm_dependencies() {
    echo "Updating helm dependencies..."
}

install_trento_server_chart() {
    echo "Installing trento-server chart..."
}

main() {
    cmdline $ARGS
    echo "Installing trento-server on k3s..."
    install_k3s
    install_helm
    update_helm_dependencies
    install_trento_server_chart
}
main