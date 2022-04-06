# AWS VPC EXPORTER

## Requirements

- `brew install go`
- `pre-commit install`

## Testing

You can either run main.go with the following command

```bash
- promu build
- aws-vault exec <profile> -- ./aws-vpc-exporter --aws.vpc-id=<vpc id>
```

or, you can build the docker image, pass the credentials and test the image.

```bash
- docker build . -t paytmlabs/aws-vpc-exporter
- aws-vault exec <AWS_PROFILE> -- docker run -p 127.0.0.1:9223:9223 -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION paytmlabs/aws-vpc-exporter --aws.vpc-id=<vpc id>
```

Now hit the port to see the result.

```bash
curl 127.0.0.1:9223/metrics | grep "aws_vpc_subnet_available_ip_addresses"
```

Sample result:
```bash
aws_vpc_subnet_available_ip_addresses{subnet_id="subnet-AAAA"} 875
aws_vpc_subnet_available_ip_addresses{subnet_id="subnet-BBBB"} 792
aws_vpc_subnet_available_ip_addresses{subnet_id="subnet-CCCC"} 501
aws_vpc_subnet_available_ip_addresses{subnet_id="subnet-DDDD"} 500
```

## Debug Logging

aws-vpc-exporter includes structured logging via `promlog`. Defaults are suitable for running the exporter in production.


To enable full debug logging (**WARNING: very verbose!**) pass the following flags when starting the service: `--log.level=debug --log.format=json`
