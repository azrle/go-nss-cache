# go-nss-cache
Fetch from the source and update nss cache.
Pair it with https://github.com/gate-sso/libnss-cache to integrate the following files to NSS databases.
```
/etc/passwd.cache
/etc/group.cache
```

And `/etc/sshkey.cache` is also generated for `AuthorizedKeysCommand` which should be set to `/path/to/authorized-keys-command`.

_NOTE: The codes are a little unnecessarily complicated. They could be simpler. This repo is only for golang programming practises._
