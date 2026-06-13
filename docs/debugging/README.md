# Obstor Server Debugging Guide

### HTTP Trace
HTTP tracing streams request trace events from running servers.

Example:
```sh
obstor server /data
```

A default trace is succinct, indicating only the API operations being called and the HTTP response status. The API and dashboard also expose verbose modes that trace the entire HTTP request, and an option to additionally include internode communication.


### Subnet Health
Subnet Health diagnostics help ensure that the underlying infrastructure that runs Obstor is configured correctly, and is functioning properly. This test is one-shot long running one, that is recommended to be run as soon as the cluster is first provisioned, and each time a failure scenario is encountered. Note that the test incurs majority of the available resources on the system. Care must be taken when using this to debug failure scenario, so as to prevent larger outages. Health tests can be triggered using Obstor's API or the dashboard.

Example:
```sh
obstor server /data
```

The health test takes no parameters. The output printed will be of the form:
```sh
● Admin Info ... ✔
● CPU ... ✔
● Disk Hardware ... ✔
● Os Info ... ✔
● Mem Info ... ✔
● Process Info ... ✔
● Config ... ✔
● Drive ... ✔
● Net ... ✔
*********************************************************************************
                                   WARNING!!
     ** THIS FILE MAY CONTAIN SENSITIVE INFORMATION ABOUT YOUR ENVIRONMENT **
     ** PLEASE INSPECT CONTENTS BEFORE SHARING IT ON ANY PUBLIC FORUM **
*********************************************************************************
Health data saved to dc-11-health_20260321053323.json.gz
```

The gzipped output contains debugging information for your system
