# drone-datadog

[![Build Status](https://cloud.drone.io/api/badges/masci/drone-datadog/status.svg)](https://cloud.drone.io/masci/drone-datadog)

This plugin lets you send events and metrics to Datadog from a drone pipeline.

## Usage

To send a metric every time a pipeline runs, add this step:

```yml
- name: count-pipeline
  image: masci/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    metrics:
      - type: "count"
        name: "masci.pipelines.count"
        value: 1.0
        tags: ["project:${DRONE_REPO_NAME}", "branch:${DRONE_BRANCH}"]
```

Sending an event is similar, both `metrics` and `events` support the `host` field:

```yml
- name: notify-pipeline
  image: masci/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Building drone-datadog success"
        text: "Version ${DRONE_TAG} is available on Docker Hub"
        alert_type: "info"
        host: ${DRONE_SYSTEM_HOSTNAME}
        priority: "low"
```

You can use events to notify something bad happened:

```yml
- name: notify-pipeline
  image: masci/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Build failure"
        text: "Build ${DRONE_BUILD_NUMBER} has failed"
        alert_type: "error"
        priority: "normal"
  when:
    status:
      - failure
```

You can change the datadog site region to EU (`com` is default)

```yml
- name: notify-pipeline
  image: masci/drone-datadog
  settings:
    region: eu
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Build failure"
        text: "Build ${DRONE_BUILD_NUMBER} has failed"
        alert_type: "error"
  when:
    status:
      - failure
```

You can look at [this repo .drone.yml](.drone.yml) file for a real world example.
