#!/bin/bash

# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script automates the creation of Cloud Workstations (CW) for admins to
# build Android Automotive OS (AAOS) device images. It sets up the necessary
# infrastructure including:
#   - Enabling required Google Cloud services (Workstations, Cloud Build, Artifact Registry).
#   - Creating an Artifact Registry repository to store the development environment Docker image.
#   - Checking if there is a development environment Docker image with the specified ASFP version.
#   - Creating a Cloud Workstations cluster.
#   - Creating a Cloud Workstations configuration for admins.
#   - Creating a Cloud Workstation for admins.

# Set default globals
LOCATION="us-west1"
ASFP_VERSION="canary"
ARTIFACT_REPO="android-dev-env"
CLUSTER="aaos"
CONFIG="admin-config"
# Set the disk size to 1TB because it takes about 550GB for AAOS 15.
PD_DISK_SIZE=1000
WORKSTATION="admin-workstation"
# Set the timeout to 4 hrs to have enough time to sync: 0.5h, build a target: 1hr & build cts: 1hr
WORKSTATION_TIMEOUT=14400

#######################################
# Check vmExternalIpAccess policy.
# Arguments:
#   Project ID
#######################################
function check_external_ip_access() {
  local project=$1

  local vm_external_ip_access=$(gcloud resource-manager org-policies describe \
      "compute.vmExternalIpAccess" \
      --format="value(listPolicy.allValues)" \
      --effective \
      --project=${project})

  echo "Info: compute.vmExternalIpAccess is ${vm_external_ip_access}."
  echo "      The workstations need to access external IPs to download the code."
}

#######################################
# Check disableNestedVirtualization policy.
# Arguments:
#   Project ID
#######################################
function check_disable_nested_virtualization() {
  local project=$1

  local disable_nested_virtualization=$(gcloud resource-manager org-policies describe \
      "compute.disableNestedVirtualization" \
      --format="value(booleanPolicy.enforced)" \
      --project=${project})

  echo "Info: compute.disableNestedVirtualization is ${disable_nested_virtualization}."
  echo "      The workstations need to run an Android virtual devic as the nested virtualization"
}

#######################################
# Check default subnet mode.
# Arguments:
#   Project ID
#######################################
function check_default_subnet_mode() {
  local project=$1

  local default_subnet_mode=$(gcloud compute networks list \
      --filter="name=( 'default' )" \
      --format="value(SUBNET_MODE)" \
      --project=${PROJECT})

  echo "Info: the default SUBNET_MODE is ${default_subnet_mode}."
  echo "      The workstations expect the default subnet mode is AUTO."
}


#######################################
# Enable required services to create an Artifact Registery repository
# Arguments:
#   Project ID, repo name, repo location
#######################################
function create_repo() {
  local project=$1
  local artifacts_repo=$2
  local location=$3

  local project_number=$(gcloud projects describe ${project} --format="value(projectNumber)")
  local service_account="${project_number}-compute@developer.gserviceaccount.com"

  # Setting up permissions to enable Cloud Build to build and push docker images.
  gcloud services enable workstations.googleapis.com --project=${project}
  gcloud services enable cloudbuild.googleapis.com --project=${project}
  gcloud services enable artifactregistry.googleapis.com --project=${project}

  echo "Enabling Cloud Build to push Docker images"
  # https://cloud.google.com/build/docs/build-push-docker-image#before-you-begin"
  gcloud projects add-iam-policy-binding ${project} \
      --member=serviceAccount:${service_account} \
      --role="roles/storage.objectUser"

  gcloud projects add-iam-policy-binding ${project} \
      --member=serviceAccount:${service_account} \
      --role="roles/artifactregistry.writer"

  gcloud iam service-accounts add-iam-policy-binding ${service_account} \
      --member=serviceAccount:${service_account} \
      --role="roles/iam.serviceAccountUser" \
      --project=${project}

  echo "Creating an repository for development environment Docker images"
  gcloud artifacts repositories create ${artifacts_repo} \
      --project=${project} \
      --repository-format=docker \
      --location=${location} \
      --description="${artifacts_repo}"
}

#######################################
# Prints script usage to stderr.
#######################################
function _print_usage() {
  (
    echo "Create adminstrator workstation for prebaking Android device build targets."
    echo
    echo "usage: $(basename $0) [OPTIONS]"
    echo "  options:"
    echo "    -p --project       Porject ID. Must specify, e.g. "
    echo "                       my-unique-cloud-project-id"
    echo "    -l --location      Cloud location. Defaults to "
    echo "                       us-west1"
    echo "    -v --asfp_version  Android Studio for Platform Version. Defaults to:"
    echo "                       canary"
    echo "    -a --artifact_repo Artifact Registry Repository. Defaults to:"
    echo "                       cloud-ide"
    echo "    -c --cluster       Cloud Workstations(CW) cluster name. Defaults to:"
    echo "                       aaos"
    echo "    -f --config        Admins CW config name. Defaults to:"
    echo "                       admin-config"
    echo "    -s --pd_disk_size  Persistent Disk size in GB. Defaults to:"
    echo "                       500"
    echo "    -w --workstation   Admins CW name. Defaults to:"
    echo "                       admin-workstation"
    echo "    -h --help          Print usage."
  ) 1>&2
}

#######################################
# Create Admins Workstations to prebake device build targets.
#######################################
function main() {
  while getopts 'p:l:v:a:c:f:s:w:h-:' arg; do
    case "${arg}" in
      -)
        case "$OPTARG" in
          project) PROJECT="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          project=*) PROJECT="${OPTARG#*=}";;
          location) LOCATION="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          location=*) LOCATION="${OPTARG#*=}";;
          asfp_version) ASFP_VERSION="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          asfp_version=*) ASFP_VERSION="${OPTARG#*=}";;
          artifact_repo) ARTIFACT_REPO="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          artifact_repo=*) ARTIFACT_REPO="${OPTARG#*=}";;
          cluster) CLUSTER="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          cluster=*) CLUSTER="${OPTARG#*=}";;
          config) CONFIG="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          config=*) CONFIG="${OPTARG#*=}";;
          pd_disk_size) PD_DISK_SIZE="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          pd_disk_size=*) PD_DISK_SIZE="${OPTARG#*=}";;
          workstation) WORKSTATION="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));;
          workstation=*) WORKSTATION="${OPTARG#*=}";;
          help) _print_usage; exit 0;;
          *)
            if [ "$OPTERR" = 1 ] && [ "${OPTSPEC:0:1}" != ":" ]; then
              cli_errors; _print_usage; exit 1;
            fi
            ;;
        esac;;
      p) PROJECT="$OPTARG";;
      l) LOCATION="$OPTARG";;
      v) ASFP_VERSION="$OPTARG";;
      a) ARTIFACT_REPO="$OPTARG";;
      c) CLUSTER="$OPTARG";;
      f) CONFIG="$OPTARG";;
      s) PD_DISK_SIZE="$OPTARG";;
      w) WORKSTATION="$OPTARG";;
      h) _print_usage; exit 0;;
      *) cli_errors; exit 1;
    esac
  done

  # Check if arguments are all exist.
  local arg_array=('PROJECT' 'LOCATION' 'ASFP_VERSION' 'ARTIFACT_REPO')
  arg_array+=('CLUSTER' 'CONFIG' 'PD_DISK_SIZE' 'WORKSTATION')

  echo "The arguments are:"
  local missing_arguments=0
  for variable in "${arg_array[@]}"; do
    if [[ -z "${!variable}" ]]; then
      echo "Error: missing ${variable}"
      missing_arguments=1
    else
      echo "${variable}=${!variable}"
    fi
  done
  if [[ ${missing_arguments} != 0 ]]; then
    echo "Error: not all required arguments are present"
    _print_usage
    exit 1
  fi

  echo
  echo "0. Checking the external IP access and disable nested virtualization for ${PROJECT}."
  check_external_ip_access ${PROJECT}
  check_disable_nested_virtualization ${PROJECT}
  check_default_subnet_mode ${PROJECT}

  local dev_img_name="${LOCATION}-docker.pkg.dev/${PROJECT}/${ARTIFACT_REPO}/android-studio-for-platform-${ASFP_VERSION}"
  echo "Info: the android-studio-for-platform container image is needed to create the admin workstation."
  echo "      Describing the expected image: ${dev_img_name}"
  gcloud artifacts docker images describe ${dev_img_name} --project=${PROJECT}
  echo "      If the image is not found, you can run the following command to build it:"
  echo "      gcloud builds submit --substitutions=_IMAGE_NAME=${dev_img_name},_ASFP_VERSION=${ASFP_VERSION} --project=${PROJECT}"

  echo
  echo "1. Creating an Artifact Registry repository"
  local repo=$(gcloud artifacts repositories describe ${ARTIFACT_REPO} --project=${PROJECT} --location=${LOCATION} --project=${PROJECT} --format="value(name)")
  local expected_repo="projects/${PROJECT}/locations/${LOCATION}/repositories/${ARTIFACT_REPO}"
  if [[ ${repo} == ${expected_repo} ]]; then
    echo "${repo} already exists."
  else
    create_repo ${PROJECT} ${ARTIFACT_REPO} ${LOCATION}
  fi

  echo
  echo "2. Creating a Workstations Cluster."
  cluster=$(gcloud workstations clusters describe ${CLUSTER} --project=${PROJECT} --region=${LOCATION} --format="value(name)")
  expected_cluster="projects/${PROJECT}/locations/${LOCATION}/workstationClusters/${CLUSTER}"
  if [[ ${cluster} == ${expected_cluster} ]]; then
    echo "${cluster} already exists."
  else
    gcloud workstations clusters create ${CLUSTER} --region=${LOCATION} --project=${PROJECT}
  fi

  echo
  echo "3. Creating an Admins Workstations Config"
  local admins_config=$(gcloud workstations configs describe  ${CONFIG} \
      --project=${PROJECT} --cluster=${CLUSTER} \
      --region=${LOCATION} --format="value(name) ")
  local expected_admins_config="${expected_cluster}/workstationConfigs/${CONFIG}"
  if [[ ${admins_config} == ${expected_admins_config} ]]; then
    echo "${admins_config} already exists."
  else
    local project_number=$(gcloud projects describe ${PROJECT} --format="value(projectNumber)")
    local service_account="${project_number}-compute@developer.gserviceaccount.com"
    gcloud workstations configs create ${CONFIG} \
        --project=${PROJECT} \
        --region=${LOCATION} \
        --cluster=${CLUSTER} \
        --machine-type=n1-standard-96 \
        --enable-nested-virtualization \
        --enable-ssh-to-vm \
        --container-custom-image=${dev_img_name}:latest \
        --service-account=${service_account} \
        --pd-disk-type=pd-ssd \
        --pd-disk-size=${PD_DISK_SIZE} \
        --running-timeout=${WORKSTATION_TIMEOUT}
  fi

  echo
  echo "4. Creating an Admins Wrokstation"
  local admins_workstations=$(gcloud workstations describe ${WORKSTATION} \
      --project=${PROJECT} \
      --config=${CONFIG} \
      --cluster=${CLUSTER} \
      --region=${LOCATION} --format="value(name)")
  local expected_admins_workstations="${expected_admins_config}/workstations/${WORKSTATION}"
  if [[ ${admins_workstations} == ${expected_admins_workstations} ]]; then
    echo "${admins_workstations} already exists."
  else
    gcloud workstations create ${WORKSTATION} \
        --project=${PROJECT} \
        --cluster=${CLUSTER} \
        --config=${CONFIG} \
        --region=${LOCATION}
  fi

  echo
  echo "After the workstation is started for the fist time"
  echo "    this'll get the source disk from the admin workstation: ${WORKSTATION}"
  local source_disk=$(gcloud compute disks list \
    --filter="labels.google-devops-environments-assigned-environment:${WORKSTATION}" \
    --format="value(name)" \
    --project=${PROJECT})

  echo "After pre-baking is done, you can create a snapshot of the source disk,"
  echo "e.g. to run the following command:"
  echo "gcloud compute snapshots create vcar-cvd-cts-$(date +"%Y%m%d") \\"
  echo "    --source-disk=${source_disk} \\"
  echo "    --source-disk-region=${LOCATION} \\"
  echo "    --storage-location=${LOCATION} \\"
  echo "    --project=${PROJECT}"
}

main "$@"