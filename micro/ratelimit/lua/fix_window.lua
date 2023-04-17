-- 先找找有没有这个限流对象的设置
local val = redis.call('get', KEYS[1])
local expiration = ARGV[1]
local limit = tonumber(ARGV[2])
if val == false then
    if limit < 1 then
        -- 执行限流
        return "true"
    else
        -- set your_service 1 px 100s
        redis.call('set', KEYS[1], 1, 'PX', expiration)
        -- 不执行限流
        return "false"
    end
elseif tonumber(val) < limit then
    -- 有这个限流对象，但是还没到阈值
    redis.call('incr', KEYS[1])
    -- 不指定限流
    return "false"
else
    -- 执行限流
    return "true"
end