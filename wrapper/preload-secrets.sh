#!/bin/bash

# WIP: this is meant for dev purposes and is unstable for now
# TODO: figure out a way to load multiline secrets from .env in bash

exec ${args[@]}

# The name of this script
name=$(basename $0)

# The full path to this script
fullPath=$(dirname $(readlink -f $0))

# The path to the interpreter and all of the originally intended arguments
args=("$@")

# Secrets .env File
secretsFile="${SECRETS_FILE}"
if [[ -f "${secretsFile}" ]]; then
    # TODO: figure out a way to load multiline secrets from .env in bash
    echo "lala"
    # -d works only on FreeBSD / MacOS
    # export $(cat "${secretsFile}" | xargs -d '\n')
    export $(cat "${secretsFile}" | tr "\n" "\0" | xargs -0)
    # export $(cat "${secretsFile}" | xargs -L1 -0)
fi

# Determine if AWS_LAMBDA_EXEC_WRAPPER points to this layer
# This is necessary as the Secret may have not specified a 
# new layer.
# Without checking, the lambda layer may be called again.
layer_name=$(basename ${AWS_LAMBDA_EXEC_WRAPPER})
if [[ "${layer_name}" == "${name}" ]]; then
    echo "No new layer was specified, unsetting AWS_LAMBDA_EXEC_WRAPPER"
    unset AWS_LAMBDA_EXEC_WRAPPER
else
    # Set args to include the new layer
    args=("${AWS_LAMBDA_EXEC_WRAPPER}" "${args[@]}")
fi

# Execute the next step
exec ${args[@]}