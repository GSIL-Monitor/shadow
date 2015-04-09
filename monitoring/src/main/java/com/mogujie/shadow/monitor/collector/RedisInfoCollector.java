package com.mogujie.shadow.monitor.collector;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import redis.clients.jedis.Jedis;

import com.mogujie.shadow.monitor.model.RedisInstanceInfo;
import com.mogujie.shadow.monitor.model.ShadowRedisInfo;
import com.mogujie.shadow.monitor.utils.RedisUtils;

/**
 * @author ziyuan
 * @date 2/5/15 10:20 AM
 * @project shadow-monitor
 */
public class RedisInfoCollector {
    private static final Logger logger = LoggerFactory.getLogger(RedisInfoCollector.class);

    private String appName;

    public RedisInfoCollector(String appName) {
        this.appName = appName;
    }

    public ShadowRedisInfo collect(RedisInstanceInfo info) {
        RedisInstanceInfo redisInstanceInfo = info;
        Jedis jedis = new Jedis(redisInstanceInfo.getIp(), redisInstanceInfo.getPort());
        String redisStatus = "";
        try {
            redisStatus = jedis.info();
        } catch (Exception e) {
            logger.error("", e);
        } finally {
            jedis.close();
        }
        ShadowRedisInfo redisInfo = RedisUtils.filterRedisStatus(redisStatus, appName);
        redisInfo.setIp(redisInstanceInfo.getIp());
        redisInfo.setPort(redisInstanceInfo.getPort());

        return redisInfo;
    }
}