# sould

**sould** is the scalable failover service for mirroring git repositories.

It is very convenient when you use some, for example, config templates in git
repositories and wants get access to your files at any time.

But how sould creates a mirrors for my repositories? Good question, my friend.

For first, sould nothing knows about your repositories and you should talk him
about it. And all he needs it is mirror name and clone url (origin).

You can create a push-receive hook which will send mirror name and clone url to
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

## Arbitary example

Jack have a git repository with templates for configs for some soft in the
repository `https://git.in.local/jack/gunter-configs` and Jack know,
`git.in.local` - it is the point of failure, because if `git.in.local` put
down, all of Jack containers will not be able to get the configs, so all Jack
infrastructure will be put down, and Jack will be put down.

Jack...
