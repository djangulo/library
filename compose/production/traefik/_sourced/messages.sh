#!/usr/bin/env bash


message_newline() {
    echo
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
    echo -e "\e[33mWARNING\e[0m: ${@}"
}

message_error()
{
    message_newline
    echo -e "\e[31mERROR\e[0m: ${@}"
    message_newline
}

message_info()
{
    # echo -e "\e[1;34mINFO\e[033m: ${@}"
    echo -e "\e[1;34mINFO\e[0m: ${@}"
}

message_suggestion()
{
    echo -e "\e[33mSUGGESTION\e[0m: ${@}"
}

message_success()
{
    echo -e "\e[32mSUCCESS\e[0m: ${@}"
}
