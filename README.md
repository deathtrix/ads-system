# Ads System POC


### RUN
- docker-compose up
- navigate to:   http://localhost:8080/


### TODO
- DSP:
    - /status/{id} - show user sync data
    - /sync-details - get audience_details and merge them in redis if timestamp is newer
        - use LUA to check for timestamps in the same transaction and use RedisCluster to distribute (locks supported)
    - /add-ssp - POST - add new SSP to 

- SSP:
    - /sync-url endpoint
        - if dsp_cookie_id in redis -> return ssp_cookie_id
        - if not in redis -> generate it and save to redis
        - if {resync}:
            - return 301 http://DSP/{sync-url} with cookie - myssp_id={id}
        - else:
            - return 200 with cookie - myssp_id={id}
    - /status/{id} - show user sync data
    - /sync-details - get audience_details and merge them in redis if timestamp is newer
        - use LUA to check for timestamps in the same transaction and use RedisCluster to distribute (locks supported)

### Improvements
- refactor
    - redis prefixes as constants
    - extract methods
- authentication server-server communication - ex: Oauth2 client_credentials grant
- use kafka streams for profile sync - async rather than sync


### Use
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
