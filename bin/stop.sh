#!/bin/bash
PID_FILE=$PWD/bin/.pid
PID=$(cat $PID_FILE)

if kill $PID; then
    rm $PID_FILE
if
