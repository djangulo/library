#!/usr/bin/env bash


SCRIPT_WARNING="\nThis script provides the basic setting for a traefik-scalable config\n(with docker), it may make changes to the machine's filesystem."
DOCKER_REQUIRED="\nThis script requires docker-machine to run: docker-machine not found in \$PATH.\n Visit https://docs.docker.com/install/overview/ for installation instructions."
DOCKER_MACHINE_REQUIRED="\nThis script requires docker-machine to run: docker-machine not found in \$PATH.\n Visit https://docs.docker.com/machine/install-machine/ for installation instructions."
DOCKER_COMPOSE_REQUIRED="\nThis script requires docker-compose to run: docker-compose not found in \$PATH.\n Visit https://docs.docker.com/compose/install/ for installation instructions."
AWS_REQUIRED="\sThis script requires aws-cli to run: aws not found in \$PATH.\nVisit https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html for installation instructions"
SED_REQUIRED="\n This script requires GNU sed <https://www.gnu.org/software/sed/>: sed not found in \$PATH"
