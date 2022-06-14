# gnmi-fake
gNMI fake server to work with [ygot getting-started example](https://github.com/openconfig/ygot/tree/master/demo/getting_started).

```shell
# Run
go run main.go -bind_address :9339 -config testdata.json -notls

# Test
go install github.com/openconfig/gnmi/cmd/gnmi_cli@latest
gnmi_cli -a :9339 -capabilities -insecure
```