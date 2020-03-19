# Ads System POC

### RUN
- docker-compose up
- navigate to:   http://localhost:8080/
- see cookie status on DSP: http://localhost:5050/status/{dsp_cookie_id}
- see cookie status on SSP: http://localhost:6060/status/{dsp_cookie_id}
- register new SSP: POST to http://localhost:5050/add-ssp (use Postman for ex)
Payload:
```{
    "name":             "ssp2",
    "sync-url":         "http://localhost:6061/sync/sync.gif",
    "sync-details-url": "http://localhost:6061/sync/usersync",
    "cookie-name":      "ssp2_cookie",
    "resync":           "0"
}
```

### TODO
- fix bug - first redirect doesnt set SSP cookie
- SSP - read DSP url from database (seeded)
- DSP and SSP:
    - /sync-details - get audience_details and merge them in redis if timestamp is newer
        - use LUA to check for timestamps in the same transaction and use RedisCluster to distribute (locks supported)

### Improvements
- refactor - redis prefixes as constants
- authentication server-server communication - ex: Oauth2 client_credentials grant
- use kafka streams for profile sync - async rather than sync

### Use
Redis audience merge/update on timestamp
```SCRIPT LOAD "
local c = tonumber(redis.call('get', KEYS[1]));
if c then 
    if tonumber(ARGV[1]) > c then 
        redis.call('set', KEYS[1], ARGV[1]) 
        return tonumber(ARGV[1]) - c 
    else 
        return 0 
    end 
else 
    return redis.call('set', KEYS[1], ARGV[1])
end"

EVALSHA "2ab979bc4b89ab15c14805586c33d898f99a53d4" 1 timestamp 245
```
