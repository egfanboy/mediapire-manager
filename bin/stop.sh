#!/bin/bash
PID_FILE=$PWD/bin/.pid
PID=$(cat $PID_FILE)

if kill $PID; then
    echo "Mediapire Manager successfully stopped"
    rm $PID_FILE
fi
