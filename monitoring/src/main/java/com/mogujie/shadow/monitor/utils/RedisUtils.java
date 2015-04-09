package com.mogujie.shadow.monitor.utils;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.mogujie.shadow.monitor.model.ShadowRedisInfo;

/**
 * @author ziyuan
 * @date 2/5/15 10:46 AM
 * @project shadow-monitor
 */
public class RedisUtils {
    private static final Logger logger = LoggerFactory.getLogger(RedisUtils.class);

    public static ShadowRedisInfo filterRedisStatus(String redisStatus, String appName) {
        ShadowRedisInfo redisInfo = new ShadowRedisInfo();

        redisInfo.setAppName(appName);

        Pattern pattern_used_memory = Pattern.compile("used_memory:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_used_memory = pattern_used_memory.matcher(redisStatus);
        if (matcher_used_memory.find()) {
            redisInfo.setUsed_memory(Long.valueOf(matcher_used_memory.group(1)));
        }

        Pattern pattern_used_memory_peak = Pattern.compile("used_memory_peak:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_used_memory_peak = pattern_used_memory_peak.matcher(redisStatus);
        if (matcher_used_memory_peak.find()) {
            redisInfo.setUsed_memory_peak(Long.valueOf(matcher_used_memory_peak.group(1)));
        }

        Pattern pattern_mem_fragmentation_ratio = Pattern.compile("fragmentation_ratio:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_mem_fragmentation_ratio = pattern_mem_fragmentation_ratio.matcher(redisStatus);
        if (matcher_mem_fragmentation_ratio.find()) {
            redisInfo.setMem_fragmentation_ratio(Double.valueOf(matcher_mem_fragmentation_ratio.group(1)));
        }

        Pattern pattern_total_commands_processed = Pattern.compile("total_commands_processed:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_total_commands_processed = pattern_total_commands_processed.matcher(redisStatus);
        if (matcher_total_commands_processed.find()) {
            redisInfo.setTotal_commands_processed(Long.valueOf(matcher_total_commands_processed.group(1)));
        }

        Pattern pattern_total_connections_received = Pattern.compile("total_connections_received:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_total_connections_received = pattern_total_connections_received.matcher(redisStatus);
        if (matcher_total_connections_received.find()) {
            redisInfo.setTotal_connections_received(Long.valueOf(matcher_total_connections_received.group(1)));
        }

        Pattern pattern_connected_clients = Pattern.compile("connected_clients:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_connected_clients = pattern_connected_clients.matcher(redisStatus);
        if (matcher_connected_clients.find()) {
            redisInfo.setConnected_clients(Long.valueOf(matcher_connected_clients.group(1)));
        }

        Pattern pattern_keyspace_hits = Pattern.compile("keyspace_hits:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_keyspace_hits = pattern_keyspace_hits.matcher(redisStatus);
        if (matcher_keyspace_hits.find()) {
            redisInfo.setKeyspace_hits(Long.valueOf(matcher_keyspace_hits.group(1)));
        }

        Pattern pattern_keyspace_misses = Pattern.compile("keyspace_misses:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_keyspace_misses = pattern_keyspace_misses.matcher(redisStatus);
        if (matcher_keyspace_misses.find()) {
            redisInfo.setKeyspace_misses(Long.valueOf(matcher_keyspace_misses.group(1)));
        }

        Pattern pattern_expired_keys = Pattern.compile("expired_keys:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_expired_keys = pattern_expired_keys.matcher(redisStatus);
        if (matcher_expired_keys.find()) {
            redisInfo.setExpired_keys(Long.valueOf(matcher_expired_keys.group(1)));
        }

        Pattern pattern_evicted_keys = Pattern.compile("evicted_keys:(\\d+\\.\\d+|\\d+)");
        Matcher matcher_evicted_keys = pattern_evicted_keys.matcher(redisStatus);
        if (matcher_evicted_keys.find()) {
            redisInfo.setEvicted_keys(Long.valueOf(matcher_evicted_keys.group(1)));
        }

        return redisInfo;
    }
}
