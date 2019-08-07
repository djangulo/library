#!/usr/bin/env bash


message_newline() {
    echo ''
}

message_debug()
{
    echo -e "DEBUG: ${@}"
}

message_welcome()
{
    echo -e "\e[1m${@}\e[0m"
}

message_warning()
{
    echo -e "WARNING: ${@}"
}

message_error()
{
    message_newline
    echo -e "ERROR: ${@}"
    message_newline
}

message_info()
{
    # echo -e "\e[1;34mINFO\e[033m: ${@}"
    echo -e "INFO: ${@}"
}

message_suggestion()
{
    echo -e "SUGGESTION: ${@}"
}

message_success()
{
    echo -e "SUCCESS: ${@}"
}
