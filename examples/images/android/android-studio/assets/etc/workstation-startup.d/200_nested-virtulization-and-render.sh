#!/bin/bash
# TODO: remove this after we start setting this in the base image.
echo "Setting nested virtualization and render permissions."

if [[ $(getent group kvm) ]]; then
  usermod -aG kvm user
fi

if [[ $(getent group render) ]]; then
  usermod -aG render user
fi
