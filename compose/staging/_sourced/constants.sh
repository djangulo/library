#!/usr/bin/env bash


SCRIPT_WARNING="This script provides the basic setting for a traefik-scalable config\n(with docker), it may make changes to the machine's filesystem."
DOCKER_REQUIRED="This script requires docker-machine to run: docker-machine not found in \$PATH.\n Visit https://docs.docker.com/install/overview/ for installation instructions."
DOCKER_MACHINE_REQUIRED="This script requires docker-machine to run: docker-machine not found in \$PATH.\n Visit https://docs.docker.com/machine/install-machine/ for installation instructions."
DOCKER_COMPOSE_REQUIRED="This script requires docker-compose to run: docker-compose not found in \$PATH.\n Visit https://docs.docker.com/compose/install/ for installation instructions."
SED_REQUIRED="This script requires GNU sed <https://www.gnu.org/software/sed/>: sed not found in \$PATH"
