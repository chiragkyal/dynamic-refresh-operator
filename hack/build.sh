#!/bin/bash
set -e

podman build -t dynamic-refresh-operator .
podman tag dynamic-refresh-operator quay.io/ckyal/dynamic-refresh-operator:latest
podman push quay.io/ckyal/dynamic-refresh-operator:latest
