# Packet Demo App

This is a simple command line application that demonstrates to deploy and terminate a bare-metal Packet device using Packer REST API service.

## Usage

**Available options:**

```
-bilcycle string
        Billing cycle (default "hourly")
  -facility string
        Datacenter facility code where to deploy device (default "ams1")
  -hostname string
        Hostname of the server to be deployed (default random string)
  -os string
        Server OS slug (default "centos_7")
  -plan string
        Server deployment plan (default "baremetal_0")
  -prid string
        project ID (default "")
  -token string
        Packet API key token (default "")
```

You must provide at least a token key and project ID as input flags or set environment variables.

```
export PACKET_AUTH_TOKEN="Your token key here"
export PACKET_PROJECT_ID="Your project ID here"
```

Clone the repository and run locally:

```
go run main.go
```