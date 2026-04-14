<!--
Copyright 2024-2026 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Android Open Source Repository Builder

Example installs the requisite tooling to the Workstations Base image needed to build [Android](https://source.android.com/) from source. Additionally, it includes a [script](assets/google/scripts/build-aosp.sh) that may be invoked to automate the build.

> Note: In order to interact with the Cuttlefish emulator, you must enable nested virtualization in the [machine settings](https://cloud.google.com/workstations/docs/create-configuration#define_machine_settings) of your Cloud Workstations configuration.

```
gcloud builds submit --substitutions _IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/aosp-repo-builder
```

Or can be built locally using:

```
docker build -t aosp-repo-builder .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 2222:22 aosp-repo-builder
```
> Note: we recommend running the command with the `--privileged` which more accurately reflects the running conditions of the container while running in Cloud Workstations. This is not strictly required if you don't need to exercise all functionality provided by Cloud Workstations. For example, omitting the flag will disable docker-in-docker functionality.

Then use an ssh client to connect to the container:

```
ssh user@localhost \
  -p 2222 \
  -o StrictHostKeyChecking=no \
  -o UserKnownHostsFile=/dev/null
```
