# Extend Preconfigured base images to use systemd

This example shows how to extend Cloud Workstations predefined base images to use `systemd` as the init system. Cloud Workstations predefined base images include a [simplified init system](https://cloud.google.com/workstations/docs/customize-container-images#cloud-workstations-base-image-structure) which is easily extended and usually sufficient, but some image customizations may introduce dependencies on systemd.

This example can be built with the included cloudbuild.yaml by specifying substitutions for the base image and the image name for the newly built systemd image:

```
gcloud builds submit --substitutions _BASE_IMAGE=us-central1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest,_IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/code-oss-systemd
```

Or can be built locally using:

```
docker build -t code-oss-systemd .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 8080:80 code-oss-systemd
```

> Note: this conainer must be started using the `--privileged` switch.

Then navigate to localhost:8080 on your local machine to access Code OSS.
