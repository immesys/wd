# Systemd monitoring watchdogs

This creates a couple watchdogs based on systemd units

```
prefix.sd.<displayname>
```

```

You configure it with parameters something like this:

```bash
/usr/bin/sdmon \
  --prefix "my.watchdog.prefix" \
  --interval 5m \
  --holdoff 10m \
  --unit btrdb \
  --unit prod-mysql:mysql
```

The interval is how often a watchdog is kicked (the timeout is double the interval). The holdoff is how long a unit must be running before it is considered KGOOD (if your app is crashing every 5 minutes, that should not count as KGOOD). Multiple --unit flags can be passed. The syntax is either `--unit servicename` where `.service` is automatically appended to the servicename, or `--unit servicename:displayname` where the watchdog is created with `displayname` instead of `servicename`. The latter syntax is useful if your unit contains characters like `-` that are not permitted in watchdog names.
