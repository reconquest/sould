# sould

![soul stone](https://cloud.githubusercontent.com/assets/8445924/9289433/29a8d22c-435f-11e5-9be8-294c41670873.png)

**sould** is the scalable failover service for mirroring git repositories.

It can be very useful if you store your files in the git repositories and want
to get access to them at any time.

But how sould creates a mirrors for repositories? Let me explain this.

At first, sould knows nothing about your repositories, so you should talk him
about it. All it needs are the mirror name and clone url (the origin).

Sould uses the HTTP REST-like interface for communication, and implements two
methods: `POST` and `GET`.

## API

### POST
On this request sould server will fetch all repository changes. If the server
works in master mode, then request will be propagated to all known slaves.
Master server uses parallel threads while communicating with slaves.

Basic response statuses:

- `200 OK` - everything is okay, all changes were just replicated.
- `500 Internal Server Error` - this status is returns when sould server have
     problems with pulling changeset, i.e couldn't write to storage directory,
     or could not pull remote changeset.

Master server response statuses:

- `500 Internal Server Error` - this status is returns when master server
     successfully propagated pull request to all sould slave servers, but have
     problems with pulling remote changeset on master server.
- `502 Bad Gateway` - one or more slave servers returned error statuses.
- `503 Service Unavailable` - this is fatal error, which can occur only
    when sould servers (including master one) couldn't pull the repository
    changes or returns reports about their internal server errors.

**sould** server always reports about all occured errors to the stderr and http
response, so if master server gets error report from slave server, he will log
all reports and forward it to the end user via http response.

### GET
On this request sould server will create a tar archive with content of the
latest revision.

Error statuses:
- `500 Internal Server Error` - this status can be returned only when sould
     have some internal problems, i.e. failed while reading mirror directory.
     Error details will also be written to the http response.
- `404 Not Found` - sould server didn't know about specified mirror.

**sould** will try to pull repository changes on this request in the next
cases:
- *last pull has been failed*

    I'll give an example: a client just made push to the origin repository. The
    push-receive hooks sends update request to  the sould master server. If, at
    this moment, the master server, can't connect to the origin repository,
    it'll remember that, and later, when some client tries to get a tar
    archive, the sould server will try to make a pull again.

- *sould has been restarted*

    Anything can happen. But if sould hasn't never pull changes to this mirror
    (mirror directory can be just copied to storage directory), then it will
    try to make a pull again.

**sould** also always sends http headers:

- `X-State` - this header contains the latest pull status. It can be either
    `success` or `failed`.

- `X-Date` - date of latest successfully mirror update.

## Usage

I'll tell you a story about my imaginary friend Jack. Jack is software
engineer, as you can guess.

Jack has a git repository which stores templates of configuration files for
some software. The URL of this repository is
`https://git.in.local/jack/gunter-configs`. Suppose that this configuration
files must be rendered from their templates very frequently, i.e. once a
minute. So, server `git.in.local` which serving Jack's repository is the point
of failure.  Because if it goes down, all Jack's templates will be missed, so
Jack's software will be unworkable, so Jack's infrastructure depending on this
software will also go down, and Jack himself will be go down by it's boss, and
maybe Jack's soul will also be go down, far below ground.

So, what Jack (and anyone) should do to avoid this? He should create a failover
cluster of mirrors of his single repository. That means to do:

1. [Setup post-receive hook in repository at `git.in.local`.](#setup-post-receive-hook)
2. [Setup slave sould servers.](#setup-slave-sould-servers)
3. [Setup master sould server.](#setup-master-sould-server)

### Setup post-receive hook

To tell the sould server to pull a changes Jack (or anyone) should create the
post-receive hook which will send mirror name and clone url to sould server on
every push event.

Generally, this hook should looks like this:

```bash
#!/bin/sh

NAME="my/dev/mirror"
ORIGIN="https://git.in.local/user/jack/configs.git"

SOULD="master.sould.local"
TIMEOUT="10"

echo "sending data to sould server $SOULD"
exec curl --data "name=$NAME&origin=$ORIGIN" -m "$TIMEOUT" $SOULD
```

and should be placed inside `hooks/post-receive` on Jack's git server.

[Read more about git hooks technology](https://raw.githubusercontent.com/git/git/master/Documentation/githooks.txt)

### Setup slave sould servers

Prepare to flight, because sould is easy in configuration. Basic config looks
like this:

```
listen = ":80"
storage = "/var/sould/"
```

- `listen` directive is used to set network address which sould should listen.

- `storage` directive is used to set path to root directory for all new
    mirrors, So, in this example, when your create a mirror with the name
    `dev/configs`, sould will create the repository in
    `/var/sould/dev/configs/`.

### Setup master sould server

It's a little bit harder than slave. Master config is based on the slave one
and extends it with this directives:

- `master` - set `true` for master server and `false` for slave.

    So if you want turn off master mode, you should set `master` value to
    `false` and give signal to the sould server to reload configuration file,
    after this server will work in slave mode.

- `slaves` - this directive contain list of one or more slave servers where
    replicate request will be propagated to.

- `timeout` - timeout on the whole time for communication with a single slave
    server. Base unit is millisecond.

Example:
```
listen = ":80"
storage = "/var/sould/"
master = true
slaves = ["slave1.local", "slave2.local"]
timeout = 20000
```

A server with that configuration will be the master server, which listens at 80
port. All requests it will propagate to the servers `slave1.local` and
`slave2.local`, and it will use timeout in 20 seconds for communication with
each of them. `/var/sould`/ will be used as the storage directory.

##### Reloading configuration

**sould** will reload configuration if it catches `SIGHUP` signal, so you can
turn any slave server to master mode, and, of course, can turn any master
server to slave server.

## Running

Synopsis:
```
sould -c <config> [--unsecure]
```

- `-c <config>` - use specified config file.
- `--unsecure` -- run sould in unsecure mode. Without this flag, sould
    does not create and not update repositories where `origin` parameter is a
    some local path.
