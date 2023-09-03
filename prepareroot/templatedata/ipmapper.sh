#!/bin/bash

# Get the system's hostname
hostname=$(hostname)

# Get the system's private IPv4 address
ip=$(hostname -I | cut -d' ' -f1)

# Call curl with the provided URL
curl "http://hostsystem:8080/add?domain=$hostname&ip=$ip"
