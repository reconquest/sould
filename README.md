# sould

**sould** is the scalable failover service for mirroring git repositories.

It is very convenient when you use some, for example, config templates in git
repositories and wants get access to your files at any time.

But how sould creates a mirrors for my repositories? Good question, my friend.

For first, sould nothing knows about your repositories and you should talk him
about it. And all he needs it is mirror name and clone url (origin).

For communication sould uses the HTTP REST-like interface, and implements two
methods - `POST` and `GET`.

## API

### POST
On this request sould server will fetch all repository changes. If server works
in master server, then request will be propagated to all known slaves. In
communication with slaves, master mode uses a parallel threads.

Basic response statuses:
- `201 Created` - this status returns when sould does not know about this
repository, and he had create new mirror.
- `200 OK` - everything is okay, all changes has been replicated.
- `500 Internal Server Error` - this status returns when sould server have some
    internal problems, i.e could not write to storage directory, or could not
    pull repository changeset.

Master server response statuses:
- `502 Bad Gateway` - one or more slave servers returned error statuses.
- `503 Service Unavailable` - this is fatal error, which can be occured only
    when all sould (including master) servers could not pull repository
    changeset or returns report about internal server errors.

**sould** server always reports about all occured errors to stderr and to http
output, so if master server gets error report from slave server, he will log
all reports and forward it to the end user's http output.

### GET
On this request sould server will create a tar archive with content of latest
revision.

Error statuses:
- `500 Internal Server Error` - this status can be returned only when sould
     have a some internal problems, i.e. failed for reading mirror directory.
     Error details also should be written to http output.
- `404 Not Found` - sould server does not known about requested mirror.

**sould** will try to pull repository changes on this request in a few cases:
- *last pull has been failed*

    Practice example. Client makes push to origin repository, push-receive hook
    sends update request to sould master server. And if master server, at this
    moment, could not connect to origin repository, he remember it, and when
    some client will try to get a tar archive, sould will try to make a pull.

- *sould has been restarted*

    Anything can happen. But if sould does not pull changes already to this
    mirror (mirror directory can be just copied to storage directory), then he
    will try to make a pull request.

**sould** also always sends http headers:

- `X-State` - that header shows the latest pull status. It can be `success` or
    `failed`.

- `X-Date` - date of latest successfully mirror update.

## Usage

I want tell you story about my imaginary friend Jack. Jack is engineer, how you
can guess.

Jack have a git repository with templates for configs for some soft in the
repository `https://git.in.local/jack/gunter-configs` and Jack know,
`git.in.local` - it is the point of failure, because if `git.in.local` put
down, all of Jack containers will not be able to get the configs, so all Jack
infrastructure will be put down, and Jack will be put down, and Jack's soul
will be put down.

So, Jack should create a failover cluster of mirrors to him repository, for
that he should do:

1. [Setup post-receive hook in repository at `git.in.local`.](#setup-post-receive-hook)
2. [Setup slave sould servers.](#setup-slave-sould-servers)
3. [Setup master sould server.](#setup-master-sould-server)

### Setup post-receive hook

You should create a post-receive hook which will send mirror name and clone
url to sould server.

Generally, hook should look like this:

```bash
#!/bin/sh

NAME="my/dev/mirror"
ORIGIN="https://git.in.local/user/jack/configs.git"

SOULD="master.sould.local"
TIMEOUT="10"

echo "sending data to sould server $SOULD"
exec curl --data "name=$NAME&origin=$ORIGIN" -m "$TIMEOUT" $SOULD
```

and should locates in `hooks/post-receive` your bare repository.

[Read more about git hooks technology](https://raw.githubusercontent.com/git/git/master/Documentation/githooks.txt)

### Setup slave sould servers

Prepare to flight, because sould is easy in configuration, basic config look
like this:

```
listen = ":80"
storage = "/var/sould/"
```

- `listen` directive talks about what address sould should listen to.

- `storage` directive is a path to directory, which will be used as a root
 directory for all new mirrors, so if you wants create a mirror with name
 `dev/configs`, sould will create a *bare* repository in
 `/var/sould/dev/configs/`.

### Setup master sould server

There little bit harder then slave. Master config extends slave config with
this directives:

- `master` - that flag should be `true`, if server is master. So if you want
    turn off master mode, you should set `master` value to `false` and give
    signal to the sould server to reload configuration file, after this server
    will work in slave mode.
- `slaves` - this directive contain list of one or more slave servers where
    replicate request will be propagated to.
- `timeout` - timeout on all the time for communication with a one slave
    server. Measures in milliseconds.

Example:
```
listen = ":80"
storage = "/var/sould/"
master = true
slaves = ["slave1.local", "slave2.local"]
timeout = 20000
```

Server with given configuration will be a master server, which listen at 80
port, all replicate requests propagates to `slave1.local` and `slave2.local`,
and for communication with every slave server will use timeout 20 seconds. As
storage directory will be used `/var/sould/`.

##### Reloading configuration

**sould** reload configuration, when catch `SIGHUP` signal, so you can turn any
slave server to master mode, and, of course, can turn any master server to
slave server.

## Running

Be default, **sould** read config file from `/etc/sould.conf`, but path to
configuration file can be changed via specifying `-c <config>` argument.

Example:
```
sould -c /tmp/sould.conf
```

If you need to create local mirrors to local repositories (on the same
filesystem), then you should pass `--unsecure` flag. Without this flag, sould
does not create and not update repositories where `origin` parameter is a some
local path.
