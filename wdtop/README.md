# WatchDog top

This creates a couple watchdogs based on system parameters and maintains them.

```
prefix.cpu - cpu usage
prefix.memory - free memory
prefix.disk.<name> - free disk space on a given directory
prefix.ps.<name> - is a process running
```

You configure it with parameters something like this:

```
/usr/bin/wdtop \
  --prefix "my.watchdog.prefix" \
  --min-mem-mb 1000 \
  --max-cpu-percent 50 \
  --df /:root:2000 \
  --df /home:home:2000 \
  --interval 2m \
  --proc btrdb:db \
  --proc spawnd:spawnd
```

The syntax for df is `<dir>:<name>:<free_mb>` and the syntax for proc is `<executable name>:<display name>`.
