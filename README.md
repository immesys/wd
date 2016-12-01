
The concept of global watchdogs are simple. Every watchdog has a name, and part
of that name is used as a prefix for grouping them together, e.g:

```
myservice.us-west-1a.server.healthy
```

And all that the watchdog framework allows you to do is five operations:

- *kick* a watchdog, which resets its timer or creates it if it didn't exist
- *fault* a watchdog, which causes it to enter failed state before the timeout
- *retire* a watchdog, which deletes it
- *auth* a prefix, which means create a new key that is allowed to interact with a subset of your current key
- get the *status* of a set of watchdogs given by their prefix

To set up the authentication key, put your authentication token (in hex plaintext) in one of these places

- `$WD_TOKEN` environment variable
- `.wd_token` file in your current directory
- `.wd_token` file in your `$HOME` directory
- `/etc/wd/token` file

# Python specifics

The python bindings are very similar to the command line, you have five functions available:

```python
gwd.kick(name, timeout=300)
gwd.fault(name, reason)
gwd.retire(prefix)
key = gwd.auth(prefix)
stats = gwd.status(prefix)
```

That's it!
