# ecs2k8s 

> :warning: **This project is in development phase. Please do not use it in production environments.**


`ecs2k8s` is a CLI tool that will be able to migrate a running cluster from ECS to a Kubernetes cluster. It takes a ECS task definition and translates into corresponding Kubernetes objects, then deploying into the desired cluster.

## Usage

The usage instructions can be viewed by entering ```ecs2k8s --help```

```

    ecs2k8s - A CLI tool that will be able to migrate a running cluster from ECS to a Kubernetes cluster

    Usage:
    ecs2k8s [command]

    Available Commands:
    ecs         A set of commands to work with the existing AWS ECS cluster.
    generate    Generate the K8s definition YAML or JSON file for the tasks
    migrate     Migrate ECS cluster to the k8s cluster.
    help        Help about any command

```

### Available commands

- List ECS task definitions 

```bash
    $ ecs2k8s ecs list-tasks
```

- Generate K8s definition YAML|JSON file

```bash
    $ ecs2k8s ecs generate --task <Active-Task-Definition-Name>
```

- Create a Kubernetes deployment in a cluster, reads by default from local kube config file

```bash
    $ ecs2k8s ecs migrate --task <Active-Task-Definition-Name>
```

## Requirements

-	[Go](https://golang.org/doc/install) >= 1.16