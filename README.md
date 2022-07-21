# AWS Lambda Secrets
Load and persist AWS SecretsManager secrets to AWS Lambda Functions


## Description

The Go binary `lambda-secrets` reads SecretsManager ARNs or Paths and Environment Variable names from a YAML file and loads the contents of the ARNs to the named Environment Variables. Additionally it is capable to run a process (with the Environment Variables attached) defined in positional arguments or the `ENTRYPOINT` Environment Variable.

## YAML File syntax

The YAML file can be located anywhere and the location can be set to the binary through the `-f` flag. Default is `secrets.yaml` in the working directory.

The syntax of the file is as follows:

```yaml
secrets: # a YAML list that
	# contains maps of `valueFrom` and `name` keys
  - valueFrom: 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:lambda-secrets/test1-VeF2Fu' # <-- Full ARN - Works!
    name: SECRET_VALUE_1
  - valueFrom: 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:lambda-secrets/test2' # <-- ARN without suffix - Works!
    name: SECRET_VALUE_2
  - valueFrom: 'lambda-secrets/test3' # <-- Path for same AWS Account secrets - Works!
    name: SECRET_VALUE_3
  - valueFrom: '5up3r5ecre7p@55w0rd' # <-- The Secret itself - DOES NOT WORK!
    name: SECRET_VALUE_4
  [...]
```

The syntax is the YAML equivalent of the JSON [ECS Task Definition `secrets` field](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#secrets).


### Standalone

#### First Execution (Cold Start)
```bash
$ lambda-secrets env
2022/07/20 17:02:28 Positional Argument treated as entrypoint: [env]
2022/07/20 17:02:28 Looking for Dotenv file '/tmp/lambda-secrets.env'
2022/07/20 17:02:28 Dotenv file '/tmp/lambda-secrets.env' NOT found!
2022/07/20 17:02:28 Loading Secrets from AWS SecretsManager
2022/07/20 17:02:28 [+] Loading 'SECRET_VALUE_1' from 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:lambda-secrets/test1'
2022/07/20 17:02:28 [+] Loading 'SECRET_VALUE_2' from 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:lambda-secrets/test2'
2022/07/20 17:02:29 [+] Loading 'SECRET_VALUE_3' from 'lambda-secrets/test2'
2022/07/20 17:02:29 Passing execution to '[env]'

[...]
SECRET_VALUE_1={"username":"admin","password":"adm1n"}
SECRET_VALUE_2=Just a line
SECRET_VALUE_3=Just a line
[...]

2022/07/20 17:02:29 Execution finished
```

#### Later Execution (with `/tmp/lambda-secrets.env` already in-place)
```bash
$ lambda-secrets env
2022/07/20 17:03:30 Positional Argument treated as entrypoint: [env]
2022/07/20 17:03:30 Looking for Dotenv file '/tmp/lambda-secrets.env'
2022/07/20 17:03:30 Dotenv file '/tmp/lambda-secrets.env' found!
2022/07/20 17:03:30 Passing execution to '[env]'

[...]
SECRET_VALUE_1={"username":"admin","password":"adm1n"}
SECRET_VALUE_2=Just a line
SECRET_VALUE_3=Just a line
[...]

2022/07/20 17:03:30 Execution finished
```

## Container


```bash
$ docker build -t lambda-secrets .
$ docker run -ti \
  -e AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY \
  -e AWS_SESSION_TOKEN \
  -e ENTRYPOINT='env | grep SECRET_VALUE' \
  -v`pwd`/secrets.yaml:/app/secrets.yaml \
  lambda-secrets
2022/07/21 12:26:47 [*] Environment Variable 'ENTRYPOINT' is treated as entrypoint: env
2022/07/21 12:26:47 [*] Looking for Dotenv file '/tmp/lambda-secrets.env'
2022/07/21 12:26:47 [!] Dotenv file '/tmp/lambda-secrets.env' NOT found!
2022/07/21 12:26:47 [*] Loading Secrets from AWS SecretsManager
2022/07/21 12:26:47 [+] Loading 'SECRET_VALUE_1' from 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:lambda-secrets/test1'
2022/07/21 12:26:48 [+] Loading 'SECRET_VALUE_2' from 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:lambda-secrets/test2'
2022/07/21 12:26:48 [+] Loading 'SECRET_VALUE_3' from 'lambda-secrets/test2'
2022/07/21 12:26:48 [+] Passing execution to 'env'


SECRET_VALUE_1={"username":"admin","password":"adm1n"}
SECRET_VALUE_2=Just a line
SECRET_VALUE_3=Just a line
ENTRYPOINT='env | grep SECRET_VALUE'

2022/07/21 12:26:48 [+] Execution finished
```


## Reference

### Exit Codes

* `0`: Successful run of the Entrypoint with Environment Variable Secrets loaded
* `1`-`99`: File operation failure (file could not be found/read)
* `100`-`199`: AWS-related issues (wrong IAM Policies attached / no IAM credentials found, etc)
* `200`-`255`: Entrypoint Execution issues (Go `cmd.exec` failed)
