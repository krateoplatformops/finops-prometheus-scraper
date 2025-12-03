# FinOps Prometheus Scraper Generic
This repository is part of the wider exporting architecture for the Krateo Composable FinOps and scrapes Prometheus exporters to then upload the data to CrateDB through the [finops-database-handler](https://github.com/krateoplatformops/finops-database-handler).

For an in-depth look at the architecture and how to configure all the components, download the summary document [here](https://github.com/krateoplatformops/finops-operator-exporter/resources/Krateo_Composable_FinOps___Full.pdf).

## Summary
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Configuration](#configuration)

## Overview
This component is tasked with scraping a given Prometheus endpoint. The configuration is obtained from a file mounted inside the container in "/config/config.yaml". The scraper uploads all the data to a CrateDB instance, through the [finops-database-handler](https://github.com/krateoplatformops/finops-database-handler), as reported in the database-config field.

## Architecture
![Krateo Composable FinOps Prometheus Scraper Generic](resources/images/KCF-scraper.png)

## Configuration
This container is automatically started by the FinOps Operator Scraper and you do not need to install it manually.

To build the executable: 
```
make build REPO=<your-registry-here>
```

To build and push the Docker images:
```
make container REPO=<your-registry-here>
```

### Dependencies
There is the need to have an active CrateDB database, with the Krateo Database Service installed. Its login details must be placed in the database-config CR:
```yaml
apiVersion: finops.krateo.io/v1
kind: DatabaseConfig
metadata:
  name: # DatabaseConfig name
  namespace: # DatabaseConfig namespace
spec:
  username: # username string
  passwordSecretRef: # object reference to secret with password
    name: # secret name
    namespace: # secret namespace
    key: # secret key
```