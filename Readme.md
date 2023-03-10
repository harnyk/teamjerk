# Teamjerk - an unofficial Teamwork client

## What is Teamjerk?

Teamjerk is a command line client for Teamwork.com.
It is written in Go and uses the Teamwork API.

## Installation

### From source

```shell
    go install github.com/harnyk/teamjerk/cmd/teamjerk
```

### From binary

Download the latest release from [here](https://github.com/harnyk/teamjerk/releases).

### Using [eget](https://github.com/zyedidia/eget) (recommended)

Read the [eget](https://github.com/zyedidia/eget) manual for more information.

In order to take all advantages of using `eget` package manager,
you should have some bin directory in your PATH
and specify it in .eget.toml file. For example:

```toml
# File: ~/.eget.toml

[global]
target = "~/bin"
```

Now you can install Teamjerk using `eget`:

```shell
    eget harnyk/teamjerk
```

## Usage

To get help, run:

```shell
    teamjerk
```

## Typical workflow

## Log in

```shell
    teamjerk login
```

## List projects / tasks

```shell
    teamjerk tasks
```

## View time report

Current month:

```shell
    teamjerk report
```

Specify month:

```shell
    teamjerk report -m 2 # February
```

Specify year:

```shell
    teamjerk report -y 2018 -m 2 # February 2018
```

## Log working time

```shell
    teamjerk log
```

Teamjerk will ask you to specify the task and the time spent.

You can also specify everything in the arguments.
For example, the following command will log 8 hours spent on task 26658918, project 548295, on 2020-02-01, starting at 09:00:

```shell
    teamjerk log \
        -u 8 \
        -s 09:00 \
        -p 548295 \
        -t 26658918 \
        -d 2020-02-01
```

To get help, run:

```shell
    teamjerk log --help
```

## Automation (lazy employee's guide)

Write the following script (runs under **ZSH**, for Bash you need to change the syntax):

```shell
#!/usr/bin/zsh

# All working days:
declare -a dates
dates=(
    2023-03-02
    2023-03-03
    2023-03-06
    2023-03-07
    2023-03-08
    2023-03-09
    2023-03-10
)

project=548295
task=26658918

for date in "${dates[@]}"; do
    teamjerk log -u 8 -s 09:00 -p $project -t $task -d $date
done

```
