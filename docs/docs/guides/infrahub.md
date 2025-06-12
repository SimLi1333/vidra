---
title: Preparing Infrahub
sidebar_position: 2
---
import Admonition from '@theme/Admonition';

<Admonition type="note" title="Note">
This guide assumes you have a running Infrahub instance and the Vidra Operator installed in your Kubernetes cluster. If you haven't set up Infrahub yet, please refer to the [Infrahub installation guide](https://docs.infrahub.app/guides/installation).
</Admonition>

To use Infrahub, you need to define a schema resembling your resources (for example, we created a `Webserver` containing `Deployment`, `Service`, and `Ingress`, and another one called `VirtualMachine`). See the [Infrahub schema documentation](https://docs.infrahub.app/topics/schema) for more information. An example schema for a `Webserver` resource is provided at https://infrahub-operator.github.io/vidra/guides/infrahub#example-schema-for-webserver.

This guide will show you how to prepare Infrahub for use with the Vidra Operator using the example of a `Webserver` resource as example.

<Admonition type="note" title="Note">
All code snippets with example in the title can be changed to your own tailored solution.
</Admonition>

<Admonition type="note" title="Note">
There is a [Demo Repo](https://github.com/infrahub-operator/infrahub-vidra-demo) with all the necessary resources to get started with Infrahub and Vidra Operator. You can fork the repo and use it as a starting point for your own Infrahub instance.
</Admonition>

## GraphQL Queries
Vidra Operator uses a GraphQL query to fetch data from Infrahub resources. Below is the query that needs to be added to Infrahub.

<Admonition type="note" title="Note">
You can add GraphQL queries to Infrahub using the Infrahub UI, the Infrahub CLI, or through [git integration](https://docs.infrahub.app/overview/versioning#integration-with-git). For reproducibility, we recommend using git integration, as demonstrated in this guide.
</Admonition>

### Query ArtifactIDs
This query is necessary because it gathers the IDs of the relevant Artifacts. `InfrahubSync` CR of Vidra Operator will only download if the checksum changed.
```graphql
query ArtifactIDs($artifactname: [String]) {
  CoreArtifact(name__values: $artifactname) {
    edges {
      node {
        id
        storage_id {
          id
        }
        checksum {
          value
        }
        name {
          value
        }
      }
    }
  }
}
```

### Example Query Webserver Details
This query is used to get `Webserver` resources and is needed in the transformator later on.
```graphql
query GetWebserver($webserver: String!) {
  KubernetesWebserver(name__value: $webserver) {
    edges {
      node {
        name {
          value
        }
        port {
          value
        }
        containerport {
          value
        }
        replicas {
          value
        }
        image {
          value
        }
        namespace {
          value
        }
        host {
          value
        }
      }
    }
  }
}
```

## Example Transformator
The transformator is a Python script that transforms the data fetched from Infrahub into Kubernetes manifests. It uses the GraphQL queries defined above to fetch the necessary data and then generates the manifests based on a YAML template stored in the same Git repository. The transformator example below is for the `Webserver` resource, but it is as generic as possible and can easily be used for other resources, as it searches for the keys obtained from the GraphQL query and creates `customizedkeyvalue` map and searches and replaces them in the YAML template. For exxample if `name` is a field in Infrahub and it has a value of `my-webserver`, the transformator will replace `name` in the YAML template with `my-webserver`. The same applies to all other fields in the Infrahub resource.

<Admonition type="note" title="Note">
To get more information about creating your own transformator, see the [Infrahub documentation](https://docs.infrahub.app/topics/transformation).
</Admonition>

```python
from typing import Dict, Any
from infrahub_sdk.transforms import InfrahubTransform
from .helperfunctions import HelperFunctions
from pathlib import Path

""" This public module provides:
- Get information from the GraphQL
- Compare the values with the default YAML templates
"""


class TransformWebserver(InfrahubTransform):
    """Transform data into a YAML string format based on a template."""

    query = "GetWebserver"

    async def transform(self, data: Dict[str, Any]) -> str:
        """Transform the input data into a string format based on a YAML template.

        Replacing values with the matching keys from the data.
        """
        currentpath = Path(__file__).resolve()
        pathfile = str(currentpath.parents[1]) + "/YAML_Templates/webserver.yaml"
        resultstring = ""

        try:
            with open(pathfile, "r") as yamlfile:
                # Filter and extract the relevant keys from the input data
                customizedkeyvalue = HelperFunctions.filternesteddict(data)
                if not customizedkeyvalue:
                    raise ValueError("No matching keys found in the input data.")

                # Iterate through each line in the YAML template
                for line in yamlfile:
                    if ":" in line:
                        lineprefix = line.split(":")
                        lineresult = HelperFunctions.process_line(
                            "".join(str(element) for element in lineprefix[1:]),
                            customizedkeyvalue,
                        )
                        resultstring += lineprefix[0] + ":" + lineresult
                    else:
                        resultstring += line

        except FileNotFoundError:
            raise FileNotFoundError("YAML template file not found.")
        except Exception as e:
            raise RuntimeError(f"An error occurred during the transformation: {e}")

        return resultstring
```

```python
from typing import Dict, Any, cast
import re


class HelperFunctions:
    """Helper functions to process nested dictionaries and lines in text."""

    singledict: Dict[str, str] = {}

    @staticmethod
    def filternesteddict(nesteddict: Dict[str, Any], key: str = "") -> Dict[str, str]:
        """Filter nested dictionaries and store the result in a global dictionary."""
        for nestedkey, value in nesteddict.items():
            # Check if dictionary is nested
            if isinstance(value, dict):
                HelperFunctions.filternesteddict(value, nestedkey)
                continue
            if isinstance(value, list) and (
                isinstance(value[0], dict) or isinstance(value[0], list)
            ):
                HelperFunctions.filternesteddict(
                    cast(Dict[str, Any], value[0]), nestedkey
                )
                continue

            # Write the key-value pair to the global single dictionary
            HelperFunctions.singledict[key.lower()] = str(value).lower()

        return HelperFunctions.singledict

    @staticmethod
    def match_key_in_line(line: str, key: str) -> bool:
        """Check if a specific key is present in the line (case-insensitive)."""
        pattern = rf"\W{re.escape(key)}\W"  # Searching for a non-word character (like -), the key word and non-word character.
        return bool(re.search(pattern, line, re.IGNORECASE))

    @staticmethod
    def process_line(line: str, customizedkeyvalue: Dict[str, Any]) -> str:
        """Process each line, replacing matching keys with values from the input data."""
        for key, value in customizedkeyvalue.items():
            if HelperFunctions.match_key_in_line(line, key):
                line = line.replace(key, value)
        return line
```

## Example YAML Template
This is an example YAML template for a `Webserver` resources. It will be used in the transformator; the values specified in Infrahub will dynamically be replaced in the template.

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: ns-namespace

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep-name
  namespace: ns-namespace
  labels:
    app: l-name
spec:
  replicas: replicas
  selector:
    matchLabels:
      app: l-name
  template:
    metadata:
      labels:
        app: l-name
    spec:
      containers:
        - name: con-name
          image: image
          ports:
            - containerPort: containerport

---
apiVersion: v1
kind: Service
metadata:
  name: svc-name
  namespace: ns-namespace
  labels:
    app: l-name
spec:
  ports:
    - port: port
      name: http
  selector:
    app: l-name

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ing-name
  namespace: ns-namespace
spec:
  rules:
    - host: host
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: svc-name
                port:
                  number: port
```

## Example `.infrahub.yml`
This is an example of how the final artifact definition for the `Webserver` resource looks. It defines the artifact name, parameters, content type, targets, and transformation function.

```yaml
queries:
  - name: GetWebserver
    file_path: "GraphQL/GetWebserver.gql"
  - name: ArtifactIDs
    file_path: "GraphQL/ArtifactIDs.gql"

schemas:
  - "Schema/service-schema.yaml"

python_transforms:
  - name: TransformWebserver
    class_name: TransformWebserver
    file_path: "python_transform/transform_webserver.py"

artifact_definitions:
  - name: "Webserver_Artifact_Definition"
    artifact_name: "Webserver_Manifest"
    parameters:
      webserver: "name__value"
    content_type: "application/yaml"
    targets: "g_webserver"
    transformation: "TransformWebserver"
```

Once the artifact definition is created by integrating the git repo with all resources, you can create the `Webserver` resource in Infrahub and add it to the target group `g_webserver`. The Vidra Operator will then use the transformator to generate the Kubernetes manifests based on the data fetched from Infrahub.

## Example Schema for Webserver

```yaml
version: "1.0"
generics:
  - name: Resource
    namespace: Kubernetes
    description: Generic Device Data
    branch: aware
    include_in_menu: false
    display_labels:
      - name__value
    order_by:
      - name__value
    uniqueness_constraints:
      - ["name__value", "namespace__value"]
    attributes:
      - name: name
        kind: Text
        description: Name of your Webservice
        order_weight: 1
      - name: namespace
        kind: Text
        description: Namespace name - Default ns-namespace
        order_weight: 2
      - name: description
        kind: Text
        description: Additional information about the Webservice
        optional: true
        order_weight: 3
nodes:
  - name: Webserver
    namespace: Kubernetes
    icon: mdi:hand-extended
    include_in_menu: true
    generate_template: true
    inherit_from:
      - KubernetesRessource
      - CoreArtifactTarget
    attributes:
      - name: port
        kind: Number
        description: The port number on which the Service is reachable
        optional: false
        regex: ^(6553[0-5]|655[0-2][0-9]|65[0-4][0-9]{2}|6[0-4][0-9]{3}|[1-9][0-9]{0,3})$  # yamllint disable-line rule:line-length
      - name: containerport
        kind: Number
        description: The port number on which the Container is reachable
        optional: false
        regex: ^(6553[0-5]|655[0-2][0-9]|65[0-4][0-9]{2}|6[0-4][0-9]{3}|[1-9][0-9]{0,3})$  # yamllint disable-line rule:line-length
      - name: replicas
        kind: Number
        description: The number of replicas of the Deployment
        optional: false
        regex: ^[1-5]$
      - name: version
        kind: Number
        description: The version of the Deployment
        optional: true
        regex: ^[1-5]$
      - name: host
        kind: Text
        description: URL to the Webserver x.iac-ba.network.garden
        read_only: true
        optional: false
        computed_attribute:
          kind: Jinja2
          jinja2_template: "{{ name__value }}.iac-ba.network.garden"
      - name: image
        kind: Dropdown
        optional: false
        choices:
          - name: httpd:latest
            description: Image for the Apache Webserver
            color: "#7f7fff"
          - name: nginx:latest
            description: Image for the Nginx Webserver
            color: "#aeeeee"
          - name: marcincuber/2048-game
            description: Image for classic 2048 game
            color: "#008000"
          - name: public.ecr.aws/pahudnet/nyancat-docker-image
            description: Image for Nyan Cat Docker image
            color: "#FFFF00"
```