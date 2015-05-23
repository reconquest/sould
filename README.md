# sould

**sould** is a simple service which provides HTTP API for access to git
repositories.

It is very convenient for using within containers. With **sould** you can
create repositories, change and store data without using `git` utility.

## API

### Creating a repository

Generally to create a repository you need make a `POST` request to `/` with
repository name in request body.

- Repository name should be unique.
- Repository name may contain directory separators.

#### Request params
- `name` - specify repository name.

#### Response statuses
- `201 Created` - repository created.
- `209 Conflict` - repository name does not unique.

##### Example

```
curl --data "name=images/dev/new-environment" -X POST http://soul.d:80/
```

If everything is ok, sould should return `201 Created` and create a
git repository in `/var/sould/images/dev/new-environment/`.

### Creating/updating files in a repository

For creating or updating files in a repository you need to send `PUT` request
with `multipart/form-data` content type encoding and specify additional
parameter `op` with value `update` (no matter the file creates/updates
`op` should be `update`)

- Files should be stored in request body with `files[]` name.
- File names should not contains `../` and should not has prefix `.git`.

#### Response statuses
- `200 OK` - files changed.
- `404 Not Found` - repository not found.
- `403 Forbidden` - file have unsafe or malicious name.
  (file name should not contains `../` and should not have prefix `.git`)

####
