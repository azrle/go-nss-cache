# go-nss-cache
Fetch from the source and update nss cache.
Pair it with https://github.com/gate-sso/libnss-cache to integrate the following files to NSS databases.
```
/etc/passwd.cache
/etc/group.cache
```

And `/etc/sshkeys.cache` is also generated for `AuthorizedKeysCommand` which should be set to `/path/to/authorized-keys-command`.
