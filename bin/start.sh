#!/bin/bash
source $PWD/bin/env.sh

if [ -z "$MANAGER_CONFIG_PATH" ];
then
    echo "MANAGER_CONFIG_PATH not defined, please add it to the bin/env.sh file and retry again."
    exit 1
fi

if [ -z "$MANAGER_LOG_PATH" ];
then
    echo "MANAGER_LOG_PATH not defined, please add it to the bin/env.sh file and retry again."
    exit 1
fi

$PWD/bin/build.sh
echo "Starting service"
nohup ./dist/mediapire-manager --config $MANAGER_CONFIG_PATH &> $MANAGER_LOG_PATH/logs.txt &
echo "Service started with PID $!"
echo $! > $PWD/bin/.pid