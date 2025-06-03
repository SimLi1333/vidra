---
title: Preparing Infrahub
sidebar_position: 2
---
import Admonition from '@theme/Admonition';

<Admonition type="note" title="Note">
This guide assumes you have a running Infrahub instance and the Vidra Operator installed in your Kubernetes cluster. If you haven't set up Infrahub yet, please refer to the [Infrahub installation guide](https://docs.infrahub.app/guides/installation).
</Admonition>

To use Infrahub, you need to define a schema resembling your resources (we created `Webserver` containing `Deployment`, `Service` and `Ingres` and `VirtualMachine`). See the [Infrahub schema documentation](https://docs.infrahub.app/topics/schema) for more information.

This guide will show you how to prepare Infrahub for use with the Vidra Operator on the example of a `Webserver` resource. 

<Admonition type="note" title="Note">
There will be a Demo Repo with all the necessary resources to get started with Infrahub and Vidra Operator. You will be able to fork the repo and use it as a starting point for your own Infrahub instance.
</Admonition>

## GraphQL Queries
Vidra Operator uses one GraphQL query to fetch the necessary ID's of the relevant Artifacts. Below is the query which needs to be added to Infrahub. 

<Admonition type="note" title="Note">
GraphQL queries can be added to Infrahub directly via the Infrahub UI or using the Infrahub CLI or added via [git integration](https://docs.infrahub.app/overview/versioning#integration-with-git). The Example is set up to work with git integration.
</Admonition>

### Query ArtifactIDs

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
The following query is used to get `Webserver`resources and is needed in the transormator later on.

### Example Query Webserver Details

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
The transformator is a Python script that transforms the data fetched from Infrahub into Kubernetes manifests. It uses the GraphQL queries defined above to fetch the necessary data and then generates the manifests.

```python
from typing import Dict, Any
from infrahub_sdk.transforms import InfrahubTransform
from .helperfunctions import HelperFunctions
from pathlib import Path

""" This Public Module provides:
- Get Information from the GraphQL
- Compare the Values with the Default YAML Templates
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
            # Check if Dictionary is nested
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
        pattern = rf"\W{re.escape(key)}\W"  # Searching for a non-word Character (like -), de key word and non-word character.
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
This is an example YAML template for a `Webserver` resource. It will be used in the transformator the values specified in the Infrahub resource will be replaced in the template.

```YAML
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

## Example Artifact Definition
This is an example of how to the final artifact definition for the `Webserver` resource looks like. It defines the artifact name, parameters, content type, targets, and transformation function.

```yaml
artifact_definitions:
  - name: "Webserver_Artifact_Definition"
    artifact_name: "Webserver_Manifest"
    parameters:
      webserver: "name__value"
    content_type: "application/yaml"
    targets: "g_webserver"
    transformation: "TransformWebserver"
```

Once the Artifact Definition is created, you can create the `Webserver` resource in Infrahub and add it to the target group `g_webserver`. The Vidra Operator will then use the transformator to generate the Kubernetes manifests based on the data fetched from Infrahub.