# sould

**sould** is the scalable failover service for mirroring git repositories.

It is very convenient when you use some, for example, config templates in git
repositories and wants get access to your files at any time.

But how sould creates a mirrors for my repositories? Good question, my friend.

For first, sould nothing knows about your repositories and you should talk him
about it. And all he needs it is mirror name and clone url (origin).

## API

For communication sould uses the HTTP REST-like interface, and implements two
methods:

1. POST - send replicate request, on this request sould server will fetch all
repository changes. If server works in master mode, then request will be
propagated to all known slaves.

2. GET - get tar archive request, on this request sould server will create
a tar archive with content of last revision.

### Send update request

For sending update request client should send POST to `/` path on sould server.

You can create a post-receive hook which will send mirror name and clone url to
sould server, and sould, for its part, will make a clone of specified
repository, but if he already have repository with this one mirror name and
clone url, he will make a pull changes. Stupid as a fish, yep.

*sould* is very simple in configuration, basic config looks like this:
```
listen = ":80"
storage = "/var/sould/"
```

- `listen` directive talks about what address sould should listen to.

- `storage` directive is a path to directory, which will be used as a root
 directory for all new mirrors, so if you wants create a mirror with name
 'dev/configs', sould will create a *bare* repository in
 `/var/sould/dev/configs/`.

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

1. Setup post-receive hook in repository on `git.in.local`.
3. Setup slave sould servers.
2. Setup master sould server.


### Setup post-receive hook

Generally, post-receive hook should look like this:

```bash
#!/bin/sh

NAME="my/dev/mirror"
ORIGIN="https://git.in.local/user/jack/configs.git"

SOULD="master.sould.local"
TIMEOUT="10"

echo "sending data to sould server $SOULD"
exec curl --data "name=$NAME&origin=$ORIGIN" -m "$TIMEOUT" $SOULD
```

and should be located in `hooks/post-receive` your bare repository.

[Read more about git hooks technology](https://raw.githubusercontent.com/git/git/master/Documentation/githooks.txt)

### Setup slave sould servers

As arbitary example, setups
